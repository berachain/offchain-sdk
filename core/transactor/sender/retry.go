package sender

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

const (
	maxRetriesPerTx   = 3                      // TODO: read from config.
	backoffStart      = 500 * time.Millisecond // TODO: read from config.
	backoffMultiplier = 2
	maxBackoff        = 1 * time.Minute
	jitterRange       = 1000
)

// A RetryPolicy is used to determine if a transaction should be retried and how long to wait
// before retrying again.
type RetryPolicy func(context.Context, *coretypes.Transaction, error) (bool, time.Duration)

// NoRetryPolicy does not retry transactions.
func NoRetryPolicy(context.Context, *coretypes.Transaction, error) (bool, time.Duration) {
	return false, backoffStart
}

// NewExponentialRetryPolicy returns a RetryPolicy that does an exponential backoff until
// maxRetries is reached. This does not assume anything about whether the specifc tx should be
// retried.
func NewExponentialRetryPolicy() RetryPolicy {
	backoff := backoffStart
	retriesMu := &sync.Mutex{}
	retries := make(map[common.Hash]int)

	return func(ctx context.Context, tx *coretypes.Transaction, err error) (bool, time.Duration) {
		retriesMu.Lock()
		defer retriesMu.Unlock()

		txHash := tx.Hash()
		if retries[txHash] >= maxRetriesPerTx {
			delete(retries, txHash)
			return NoRetryPolicy(ctx, tx, err)
		}
		retries[txHash]++

		// Exponential backoff with jitter.
		jitter, _ := rand.Int(rand.Reader, big.NewInt(jitterRange))
		if jitter == nil {
			jitter = new(big.Int)
		}

		waitTime := backoff + time.Duration(jitter.Int64())*time.Millisecond
		if backoff *= backoffMultiplier; backoff > maxBackoff {
			backoff = maxBackoff
		}

		return true, waitTime
	}
}
