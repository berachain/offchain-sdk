package sender

import (
	"context"
	"math/big"

	sdk "github.com/berachain/offchain-sdk/types"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

var (
	multiplier = big.NewInt(11500) //nolint:gomnd // its okay.
	quotient   = big.NewInt(10000) //nolint:gomnd // its okay.
)

// TxReplacementPolicy is a function type that takes a transaction and returns a replacement
// transaction.
type TxReplacementPolicy func(context.Context, *coretypes.Transaction) *coretypes.Transaction

// DefaultTxReplacementPolicy is the default transaction replacement policy. It bumps the gas price
// by 15% (only 10% is required but we add a buffer to be safe) and generates a replacement 1559
// dynamic fee transaction.
func DefaultTxReplacementPolicy(
	ctx context.Context, tx *coretypes.Transaction,
) *coretypes.Transaction {
	sdk.UnwrapContext(ctx).Logger().Warn("processing replacement tx", "tx_hash", tx.Hash())

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
