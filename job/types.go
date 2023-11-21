package job

import (
	"context"
	"time"

	jobtypes "github.com/berachain/offchain-sdk/job/types"

	"github.com/ethereum/go-ethereum"
)

// WrapJob wraps a basic job into a job that can be submitted to the worker pool.
func WrapJob(j Basic) HasProducer {
	var wrappedJob HasProducer
	if prodJob, ok := j.(HasProducer); ok {
		wrappedJob = prodJob
	} else if condJob, ok := j.(Conditional); ok { //nolint:govet // can't avoid.
		wrappedJob = WrapConditional(condJob)
	} else if pollJob, ok := j.(Polling); ok { //nolint:govet // can't avoid.
		wrappedJob = WrapPolling(pollJob)
	}
	return wrappedJob
}

// After Basic jobs as explained in `job.go` the SDK currently
// supports two other types of jobs, polling jobs and conditional jobs.

// ============================================
// Polling Jobs
// ============================================

// Polling represents a polling job. Polling jobs are jobs that are run
// periodically at a given interval.
type Polling interface {
	Basic
	IntervalTime(ctx context.Context) time.Duration
}

// WrapPolling wraps a polling job into a conditional job, this is possible since,
// polling jobs are simply conditional jobs where `Condition()` always returns true.
// Cute little double wrap that allows us to re-use the producer from `conditional`.
func WrapPolling(c Polling) HasProducer {
	return &conditional{&polling{c}}
}

// Remember, polling is just a conditional job where the condition is always true.
var _ Conditional = (*polling)(nil)

// polling is a wrapper for a polling job.
type polling struct {
	Polling
}

// Condition always returns true for polling jobs.
func (p *polling) Condition(context.Context) bool {
	return true
}

// ============================================
// Conditional Jobs
// ============================================

// Conditional represents a conditional job.
type Conditional interface {
	Polling
	Condition(ctx context.Context) bool
}

// Wrap Conditional, wraps a conditional job to conform to the producer interface.
func WrapConditional(c Conditional) HasProducer {
	return &conditional{c}
}

// conditional is a wrapper for a conditional job.
type conditional struct {
	Conditional
}

// ConditionalProducer produces a job when the condition is met.
func (cj *conditional) Producer(ctx context.Context, pool WorkerPool) error {
	for {
		select {
		// If the context is cancelled, return.
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Check if the condition is true.
			if cj.Condition(ctx) {
				pool.SubmitAndWait(jobtypes.NewPayload(ctx, cj, nil).Execute)
			}
		}

		// NOTE: for job producers, we can just register the default, pass all of
		// them into `TaskGroups` have the context be shared and bobs your uncle
		// we get all this for free.
		// Sleep for a period of time.
		time.Sleep(cj.IntervalTime(ctx))
	}
}

// Subscribable represents a subscribable job.
type Subscribable interface {
	Basic
	Subscribe(ctx context.Context) chan any
}

// EthSubscribable represents a subscription to an ethereum event.
type EthSubscribable interface {
	Basic
	Subscribe(ctx context.Context) (ethereum.Subscription, chan any, error)
	Unsubscribe(ctx context.Context)
}
