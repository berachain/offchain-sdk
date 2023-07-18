package baseapp

import (
	"context"
	"time"

	"cosmossdk.io/log"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/worker"
)

type JobManager struct {
	// logger is the logger for the baseapp
	logger log.Logger

	// list of jobs
	jobs []job.Basic

	// worker pool
	executionPool worker.Pool
}

// New creates a new baseapp.
func NewJobManager(
	name string,
	logger log.Logger,
	jobs []job.Basic,
) *JobManager {
	return &JobManager{
		logger: logger,
		jobs:   jobs,
		executionPool: worker.NewPool(
			name+"-execution",
			4, //nolint:gomnd // hardcode 4 workers for now
			logger,
		),
	}
}

// Start.
//
//nolint:gocognit // todo: fix.
func (jm *JobManager) Start(ctx context.Context) {
	for _, j := range jm.jobs {
		if condJob, ok := j.(job.Conditional); ok {
			go func() {
				for {
					time.Sleep(50 * time.Millisecond) //nolint:gomnd // fix.
					if condJob.Condition(ctx) {
						jm.executionPool.AddTask(job.NewExecutor(ctx, condJob, nil))
						return
					}
				}
			}()
		} else if subJob, ok := j.(job.Subscribable); ok { //nolint:govet // todo fix.
			go func() {
				for {
					ch := subJob.Subscribe(ctx)
					val := <-ch
					switch val {
					case nil:
						continue
					default:
						jm.executionPool.AddTask(job.NewExecutor(ctx, subJob, val))
					}
				}
			}()
		} else if ethSubJob, ok := j.(job.EthSubscribable); ok { //nolint:govet // todo fix.
			go func() {
				sub, ch := ethSubJob.Subscribe(ctx)
				for {
					select {
					case <-ctx.Done():
						ethSubJob.Unsubscribe(ctx)
						return
					case err := <-sub.Err():
						jm.logger.Error("error in subscription", "err", err)
						// TODO: add retry mechanism
						ethSubJob.Unsubscribe(ctx)
						return
					case val := <-ch:
						jm.executionPool.AddTask(job.NewExecutor(ctx, ethSubJob, val))
						continue
					}
				}
			}()
		} else {
			panic("unknown job type")
		}
	}
}
