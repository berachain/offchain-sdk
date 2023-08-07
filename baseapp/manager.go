package baseapp

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	workertypes "github.com/berachain/offchain-sdk/job/types"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/worker"
)

// Manager handles the job and worker lifecycle.
type Manager struct {
	// logger is the logger for the baseapp
	logger log.Logger

	// list of jobs
	jobs []job.Basic

	// Job producers are a pool of workers that produce jobs. These workers
	// run in the background and produce jobs that are then consumed by the
	// job executors.
	jobProducers worker.Pool

	// Job executors are a pool of workers that execute jobs. These workers
	// are fed jobs by the job producers.
	jobExecutors worker.Pool
}

// New creates a new baseapp.
func NewManager(
	name string,
	logger log.Logger,
	jobs []job.Basic,
) *Manager {
	// TODO: read from config.
	poolCfg := worker.DefaultPoolConfig()
	poolCfg.Name = name
	poolCfg.PrometheusPrefix = "job_executor"
	return &Manager{
		logger:       logger,
		jobs:         jobs,
		jobExecutors: *worker.NewPool(poolCfg, logger),
		jobProducers: *worker.NewPool(&worker.PoolConfig{
			Name:             "job-producer",
			PrometheusPrefix: "job_producer",
			MinWorkers:       len(jobs),
			MaxWorkers:       len(jobs) + 1, // TODO: figure out why we need to +1
			ResizingStrategy: "eager",
			MaxQueuedJobs:    len(jobs),
		}, logger),
	}
}

// Start.
//

func (jm *Manager) Start(ctx context.Context) {
	for _, j := range jm.jobs {
		if sj, ok := j.(job.HasSetup); ok {
			if err := sj.Setup(ctx); err != nil {
				panic(err)
			}
		}
	}
}

// Stop.
func (jm *Manager) Stop() {
	for _, j := range jm.jobs {
		if tj, ok := j.(job.HasTeardown); ok {
			if err := tj.Teardown(); err != nil {
				panic(err)
			}
		}
	}
}

// RunProducers runs the job producers.
//
//nolint:gocognit // fix.
func (jm *Manager) RunProducers(ctx context.Context) {
	for _, j := range jm.jobs {
		// Handle migrated jobs.
		if wrappedJob := job.WrapJob(j); wrappedJob != nil {
			jm.jobProducers.Submit(
				func() {
					if err := wrappedJob.Producer(ctx, &jm.jobExecutors); err != nil {
						jm.logger.Error("error in job producer", "err", err)
					}
				},
			)
			continue
		}

		// Handle unmigrated jobs.

		if subJob, ok := j.(job.Subscribable); ok {
			jm.jobExecutors.Submit(func() {
				ch := subJob.Subscribe(ctx)
				for {
					select {
					case val := <-ch:
						jm.jobExecutors.Submit(workertypes.NewPayload(ctx, subJob, val).Execute)
					case <-ctx.Done():
						return
					default:
						continue
					}
				}
			})
		} else if ethSubJob, ok := j.(job.EthSubscribable); ok { //nolint:govet // todo fix.
			jm.jobExecutors.Submit(func() {
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
						jm.jobExecutors.Submit(workertypes.NewPayload(ctx, ethSubJob, val).Execute)
						continue
					}
				}
			})
		} else {
			panic("unknown job type")
		}
	}
}
