package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// The basic tx interface that geth core/types Transaction adheres to.
type tx interface {
	To() *common.Address
	Gas() uint64
	GasFeeCap() *big.Int
	GasTipCap() *big.Int
	Value() *big.Int
	Data() []byte
}

// NewCallMsgFromTx creates a CallMsg from a geth core/types Transaction.
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
