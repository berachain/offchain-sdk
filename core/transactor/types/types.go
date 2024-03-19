package types

import (
	"math/big"

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
