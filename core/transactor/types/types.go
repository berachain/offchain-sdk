package types

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/berachain/offchain-sdk/types/queue/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// TxRequest is a transaction request, using the go-ethereum call msg.
type TxRequest struct {
	*ethereum.CallMsg
	MsgID string // MsgID is the optional, user-provided string id for this tx request
}

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
