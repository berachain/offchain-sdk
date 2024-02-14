package types

import (
	"encoding/json"

	"github.com/berachain/offchain-sdk/types/queue/types"

	"github.com/ethereum/go-ethereum"
)

// TxRequest is a transaction request, using the go-ethereum call msg.
type TxRequest ethereum.CallMsg

// NewTxRequest returns a new TxRequest with the given type.
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
