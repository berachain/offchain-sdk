package sender

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

type (
	// Factory is an interface for building transactions, used if retrying.
	Factory interface {
		RebuildTransactionFromRequest(
			context.Context, *ethereum.CallMsg, uint64,
		) (*coretypes.Transaction, error)
	}

	// Noncer is the interface for acquiring fresh nonces, used if retrying.
	Noncer interface {
		Acquire() (uint64, bool)
	}
)

type (
	// txReplacementPolicy is a type that takes a tx and returns a replacement tx.
	txReplacementPolicy interface {
		GetNew(*coretypes.Transaction, error) (*coretypes.Transaction, error)
	}

	// retryPolicy is used to determine if a transaction should be retried and how long to wait
	// before retrying again.
	retryPolicy interface {
		Get(*coretypes.Transaction, error) (bool, time.Duration)
		UpdateTxModified(common.Hash, common.Hash)
	}
)
