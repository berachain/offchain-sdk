package sender

import (
	"math/big"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// BumpGas bumps the gas on a tx by 15% increase.
func BumpGas(tx *coretypes.Transaction) *coretypes.Transaction {
	// Bump the existing gas tip cap 15% (10% is required but we add a buffer to be safe).
	bumpedGasTipCap := new(big.Int).Mul(tx.GasTipCap(), multiplier)
	bumpedGasTipCap = new(big.Int).Quo(bumpedGasTipCap, quotient)

	// Bump the existing gas fee cap 15% (only 10% required but we add a buffer to be safe).
	bumpedGasFeeCap := new(big.Int).Mul(tx.GasFeeCap(), multiplier)
	bumpedGasFeeCap = new(big.Int).Quo(bumpedGasFeeCap, quotient)

	return coretypes.NewTx(&coretypes.DynamicFeeTx{
		ChainID:   tx.ChainId(),
		Nonce:     tx.Nonce(),
		GasTipCap: bumpedGasTipCap,
		GasFeeCap: bumpedGasFeeCap,
		Gas:       tx.Gas(),
		To:        tx.To(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	})
}

// SetNonce sets the given nonce on a tx.
func SetNonce(tx *coretypes.Transaction, nonce uint64) *coretypes.Transaction {
	return coretypes.NewTx(&coretypes.DynamicFeeTx{
		ChainID:   tx.ChainId(),
		Nonce:     nonce,
		GasTipCap: tx.GasTipCap(),
		GasFeeCap: tx.GasFeeCap(),
		Gas:       tx.Gas(),
		To:        tx.To(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	})
}
