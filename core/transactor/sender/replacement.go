package sender

import (
	"context"
	"math/big"

	sdk "github.com/berachain/offchain-sdk/types"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// TxReplacementPolicy is a function type that takes a transaction and returns a
// replacement transaction.
type TxReplacementPolicy func(ctx context.Context,
	tx *coretypes.Transaction) *coretypes.Transaction

// DefaultTxReplacementPolicy is the default transaction replacement policy.
// It bumps the gas price by 15% (only 10% is required but we add a buffer to be safe)
// and generates a replacement transaction.
func DefaultTxReplacementPolicy(
	ctx context.Context, tx *coretypes.Transaction,
) *coretypes.Transaction {
	sCtx := sdk.UnwrapContext(ctx)
	ethClient := sCtx.Chain()

	sCtx.Logger().Warn("processing replacement tx", "tx_hash", tx.Hash())

	// Get the chain to suggest a new gas price.
	newGas, err := ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil
	}

	// Bump the existing gas price 15% (only 10% required but we add a buffer to be safe).
	bumpedGasPrice := new(big.Int).Mul(tx.GasPrice(), big.NewInt(11500)) //nolint:gomnd // its okay.
	bumpedGasPrice = new(big.Int).Quo(bumpedGasPrice, big.NewInt(10000)) //nolint:gomnd // its okay.

	// Use the higher of the two.
	var gasToUse *big.Int
	if newGas.Cmp(bumpedGasPrice) > 0 {
		gasToUse = newGas
	} else {
		gasToUse = bumpedGasPrice
	}

	// Generate the replacement transaction.
	return coretypes.NewTransaction(
		tx.Nonce(), *tx.To(), tx.Value(), tx.Gas(), gasToUse, tx.Data(),
	)
}
