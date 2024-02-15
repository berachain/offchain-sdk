package types

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/berachain/offchain-sdk/types/queue/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// TxResultType represents the type of error that occurred when sending a tx.
// There are two different classes of errors:
// Errors where we fail to send it to the chain:
//
//	StatusErrSend (tx fails on sending, can retry. Usually due to nonce)
//	ERR_DECODE (tx fails on decoding, can't send to the chain)
//
// And errors where we send it to the chain, but it either never gets executed or reverts:
//
//	StatusErrReceive (tx never gets mined, probably due to low gas)
//	ERR_REVERT (tx gets mined, but reverts)
//
// We can retry on StatusErrSend and StatusErrReceive, as this is a matter of just resending the
// tx with correct values for nonce and gas.
const (
	StatusErrSend uint8 = iota
	StatusErrReceive
	StatusErrCall
	StatusRevert
	StatusSuccess
)

type (
	// TxRequest is a transaction request, using the go-ethereum call msg.
	TxRequest struct {
		*ethereum.CallMsg
		MsgID string // MsgID is the optional, user-provided string id for this tx request
	}

	// TxResult represents the error that occurred when sending a tx.
	// Nil if the tx was successful, RevertReason nil if we have an ErrSend, ErrReceive, ErrDecode.
	TxResult struct {
		Type         uint8       // always non-empty
		Error        error       // only non-empty if Type == ErrSend, ErrReceive, ErrCall
		RevertReason string      // only non-empty if Type == StatusRevert
		Hash         common.Hash // only non-empty if tx was mined
	}
)

// NewTxRequest returns a new transaction request with the given input data. The ID is optional,
// but at most 1 is allowed per tx request.
func NewTxRequest(
	to common.Address, gasLimit uint64, gasFeeCap, gasTipCap, value *big.Int, data []byte,
	msgID ...string,
) *TxRequest {
	if len(msgID) > 1 {
		panic("must only pass in 1 id for a new tx request")
	}
	return &TxRequest{
		CallMsg: &ethereum.CallMsg{
			To:        &to,
			Gas:       gasLimit,
			GasFeeCap: gasFeeCap,
			GasTipCap: gasTipCap,
			Value:     value,
			Data:      data,
		},
		MsgID: strings.Join(msgID, ""),
	}
}

// New returns a new empty TxRequest.
func (TxRequest) New() types.Marshallable {
	return &TxRequest{}
}

// String() implements fmt.Stringer.
func (tx *TxRequest) String() string {
	return tx.MsgID
}

// NewTxRequest returns a new TxRequest with the given type and error.
func (tx TxRequest) Marshal() ([]byte, error) {
	return json.Marshal(tx)
}

// Unmarshal unmarshals a TxRequest from the given data.
func (tx *TxRequest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, tx)
}
