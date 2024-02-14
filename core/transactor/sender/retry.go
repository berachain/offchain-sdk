package sender

import (
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
	backoffMultiplier = 2                      // TODO: read from config.
	maxBackoff        = 3 * time.Second        // TODO: read from config.
	jitterRange       = 1000                   // TODO: read from config.
)

// A RetryPolicy is used to determine if a transaction should be retried and how long to wait
// before retrying again.
type RetryPolicy interface {
	get(tx *coretypes.Transaction, err error) (bool, time.Duration)
	updateTxReplacement(old, new common.Hash)
}

var (
	_ RetryPolicy = (*NoRetryPolicy)(nil)
	_ RetryPolicy = (*ExpoRetryPolicy)(nil)
)

// NoRetryPolicy does not retry transactions.
type NoRetryPolicy struct{}

func (*NoRetryPolicy) get(*coretypes.Transaction, error) (bool, time.Duration) {
	return false, 0
}

func (*NoRetryPolicy) updateTxReplacement(common.Hash, common.Hash) {}

// ExpoRetryPolicy is a RetryPolicy that does an exponential backoff until maxRetries is
// reached. This does not assume anything about whether the specifc tx should be retried.
type ExpoRetryPolicy struct {
	retries sync.Map
}

func (erp *ExpoRetryPolicy) get(tx *coretypes.Transaction, err error) (bool, time.Duration) {
	var (
		txHash = tx.Hash()
		tri    *txRetryInfo
		jitter time.Duration
	)

	// If the retry error is nil, the transaction was retried successfully.
	if err == nil {
		erp.retries.Delete(txHash)
		return false, 0
	}

	txri, found := erp.retries.Load(txHash)
	if !found {
		tri = &txRetryInfo{backoff: backoffStart}
		erp.retries.Store(txHash, tri)
	} else if tri = txri.(*txRetryInfo); tri.numRetries >= maxRetriesPerTx {
		erp.retries.Delete(txHash)
		return false, 0
	}
	tri.numRetries++

	// Exponential backoff with jitter.
	if random, _ := rand.Int(rand.Reader, big.NewInt(jitterRange)); random != nil {
		jitter = time.Duration(random.Int64()) * time.Millisecond
	}
	waitTime := tri.backoff + jitter
	if tri.backoff *= backoffMultiplier; tri.backoff > maxBackoff {
		tri.backoff = maxBackoff
	}

	return true, waitTime
}

func (erp *ExpoRetryPolicy) updateTxReplacement(old, new common.Hash) {
	if txri, found := erp.retries.Load(old); found {
		erp.retries.Delete(old)
		erp.retries.Store(new, txri)
	}
}

// txRetryInfo contains the necessary information to determine if a transaction should be retried.
type txRetryInfo struct {
	numRetries int
	backoff    time.Duration
}
