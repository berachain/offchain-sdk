package sender

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Factory is an interface for building transactions, used if retrying.
type Factory interface {
	RebuildTransactionFromRequest(
		context.Context, *ethereum.CallMsg, uint64,
	) (*coretypes.Transaction, error)
	GetNextNonce(uint64) (uint64, bool)
}

type (
	// TxReplacementPolicy is a type that takes a tx and returns a replacement tx.
	TxReplacementPolicy interface {
		GetNew(*coretypes.Transaction, error) *coretypes.Transaction
	}

	// A RetryPolicy is used to determine if a transaction should be retried and how long to wait
	// before retrying again.
	RetryPolicy interface {
		Get(*coretypes.Transaction, error) (bool, time.Duration)
		UpdateTxModified(common.Hash, common.Hash)
	}
)
