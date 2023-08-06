package baseapp

import (
	"context"
	"os"

	"github.com/berachain/offchain-sdk/job"
	jobtypes "github.com/berachain/offchain-sdk/job/types"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/worker"
)

type JobManager struct {
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
func NewJobManager(
	name string,
	logger log.Logger,
	jobs []job.Basic,
) *JobManager {
	// TODO: read from config.
	poolCfg := worker.DefaultPoolConfig()
	poolCfg.Name = name
	poolCfg.PrometheusPrefix = "job_executor"
	return &JobManager{
		logger:       log.NewBlankLogger(os.Stdout),
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
//nolint:gocognit // todo: fix.
func (jm *JobManager) Start(ctx context.Context) {
	for _, j := range jm.jobs {
		if sj, ok := j.(job.Setupable); ok {
			if err := sj.Setup(ctx); err != nil {
				panic(err)
			}
		}

		if condJob, ok := j.(job.Conditional); ok { //nolint:nestif // todo:fix.
			wrappedJob := job.WrapConditional(condJob)
			jm.jobProducers.Submit(
				func() {
					if err := wrappedJob.Producer(ctx, jm.jobExecutors); err != nil {
						jm.logger.Error("error in job producer", "err", err)
					}
				},
			)
		} else if pollJob, ok := j.(job.Polling); ok { //nolint:govet // todo fix.
			wrappedJob := job.WrapPolling(pollJob)
			jm.jobProducers.Submit(
				func() {
					if err := wrappedJob.Producer(ctx, jm.jobExecutors); err != nil {
						jm.logger.Error("error in job producer", "err", err)
					}
				},
			)
		} else if subJob, ok := j.(job.Subscribable); ok { //nolint:govet // todo fix.
			jm.jobProducers.Submit(func() {
				ch := subJob.Subscribe(ctx)
				for {
					select {
					case val := <-ch:
						jm.jobExecutors.AddJob(jobtypes.NewPayload(ctx, subJob, val))
					case <-ctx.Done():
						return
					default:
						continue
					}
				}
			})
		} else if ethSubJob, ok := j.(job.EthSubscribable); ok { //nolint:govet // todo fix.
			jm.jobProducers.Submit(func() {
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
						jm.jobExecutors.AddJob(jobtypes.NewPayload(ctx, ethSubJob, val))
						continue
					}
				}
			})
		} else {
			panic("unknown job type")
		}
	}
}

// Stop.
func (jm *JobManager) Stop() {
	for _, j := range jm.jobs {
		if tj, ok := j.(job.Teardowanble); ok {
			if err := tj.Teardown(); err != nil {
				panic(err)
			}
		}
	}
}
