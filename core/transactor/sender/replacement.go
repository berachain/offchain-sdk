package sender

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/txpool"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

var _ TxReplacementPolicy = (*DefaultTxReplacementPolicy)(nil)

// DefaultTxReplacementPolicy is the default transaction replacement policy. It bumps the gas price
// by 15% (only 10% is required but we add a buffer to be safe) and generates a replacement 1559
// dynamic fee transaction.
type DefaultTxReplacementPolicy struct {
	noncer Noncer
}

func (d *DefaultTxReplacementPolicy) GetNew(
	tx *coretypes.Transaction, err error,
) (*coretypes.Transaction, error) {
	// If the sender is out of balance, return the error.
	if errors.Is(err, vm.ErrInsufficientBalance) ||
		(err != nil && strings.Contains(err.Error(), "insufficient balance for transfer")) {
		return nil, err
	}

	// Replace the nonce if the nonce was too low.
	var shouldBumpGas bool
	if errors.Is(err, core.ErrNonceTooLow) ||
		(err != nil && strings.Contains(err.Error(), "nonce too low")) {
		var newNonce uint64
		newNonce, shouldBumpGas = d.noncer.Acquire()
		tx = SetNonce(tx, newNonce)
	}

	// Bump the gas according to the replacement policy if a replacement is required.
	if shouldBumpGas || errors.Is(err, txpool.ErrReplaceUnderpriced) ||
		(err != nil && strings.Contains(err.Error(), "replacement transaction underpriced")) {
		tx = BumpGas(tx)
	}

	return tx, nil
}
