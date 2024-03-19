package types

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/berachain/offchain-sdk/types/queue/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// TxRequest is a transaction request, using the go-ethereum call msg.
type TxRequest struct {
	// CallMsg is used to provide the basic tx data. The From field is ignored for txs, only used
	// for eth calls.
	*ethereum.CallMsg

	// MsgID is the (optional) user-provided string id for this tx request.
	MsgID string

	// timeFired is the time at which this tx was initially requested; filled in automatically.
	timeFired time.Time
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
		MsgID:     strings.Join(msgID, ""),
		timeFired: time.Now(),
	}
}

// NewCallMsgFromTx creates a CallMsg from a geth Transaction. Used for rebuilding.
func NewCallMsgFromTx(t tx) *ethereum.CallMsg {
	return &ethereum.CallMsg{
		To:        t.To(),
		Gas:       t.Gas(),
		GasFeeCap: t.GasFeeCap(),
		GasTipCap: t.GasTipCap(),
		Value:     t.Value(),
		Data:      t.Data(),
	}
}

// Validate ensures that a timeFired is set on the tx request.
func (tx *TxRequest) Validate() error {
	if tx.timeFired.Equal(time.Time{}) || (tx.timeFired == time.Time{}) {
		return errors.New("timeFired must be set")
	}

	return nil
}

// Time returns the time this tx was initially requested.
func (tx *TxRequest) Time() time.Time {
	return tx.timeFired
}

// String() implements fmt.Stringer.
func (tx *TxRequest) String() string {
	return tx.MsgID
}

// New returns a new empty TxRequest.
func (TxRequest) New() types.Marshallable {
	return &TxRequest{}
}

// NewTxRequest returns a new TxRequest with the given type and error.
func (tx TxRequest) Marshal() ([]byte, error) {
	return json.Marshal(tx)
}

// Unmarshal unmarshals a TxRequest from the given data.
func (tx *TxRequest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, tx)
}

// BatchRequest is built by the factory and contains a list of TxRequests.
type BatchRequest struct {
	*coretypes.Transaction

	MsgIDs     []string
	TimesFired []time.Time
}

func (br *BatchRequest) Len() int {
	return len(br.MsgIDs)
}
