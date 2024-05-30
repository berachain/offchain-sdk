package baseapp

import (
	"crypto/rand"
	"math/big"
	"time"
)

// Retry parameters.
const (
	maxBackoff               = 2 * time.Minute
	backoffStart             = 1 * time.Second
	backoffBase              = 2
	jitterRange              = 1000
	subscriptionStaleTimeout = 1 * time.Hour
)

// withRetry is a wrapper that retries a task with exponential backoff.
func (jm *JobManager) withRetry(task func() bool) func() {
	return func() {
		backoff := backoffStart

		for {
			shouldRetry := task()
			if !shouldRetry {
				return
			}

			// Exponential backoff with jitter.
			jitter, _ := rand.Int(rand.Reader, big.NewInt(jitterRange))
			if jitter == nil {
				jitter = new(big.Int)
			}
			sleep := backoff + time.Duration(jitter.Int64())*time.Millisecond
			time.Sleep(sleep)

			backoff *= backoffBase
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}
