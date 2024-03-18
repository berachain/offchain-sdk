package sender

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/core/transactor/types"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

type (
	// TxReplacementPolicy is a type that takes a tx and returns a replacement tx.
	TxReplacementPolicy interface {
		GetNew(*coretypes.Transaction, error) *coretypes.Transaction
	}

	// Tracker is an interface for tracking sent transactions.
	Tracker interface {
		SetClient(chain eth.Client)
		Track(ctx context.Context, tx *coretypes.Transaction, msgIDs []string)
	}

	// Factory is an interface for signing transactions.
	Factory interface {
		BuildTransactionFromRequests(
			ctx context.Context, forcedNonce uint64, txReqs ...*types.TxRequest,
		) (*coretypes.Transaction, error)
		GetNextNonce(oldNonce uint64) (uint64, bool)
	}

	// A RetryPolicy is used to determine if a transaction should be retried and how long to wait
	// before retrying again.
	RetryPolicy interface {
		Get(tx *coretypes.Transaction, err error) (bool, time.Duration)
		UpdateTxModified(oldTx, newTx common.Hash)
	}
)
