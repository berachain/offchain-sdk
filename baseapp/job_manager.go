package baseapp

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	sdk "github.com/berachain/offchain-sdk/types"
	"github.com/berachain/offchain-sdk/worker"
)

const (
	producerName           = "job-producer"
	producerPromName       = "job_producer"
	producerResizeStrategy = "eager"

	executorName     = "job-executor"
	executorPromName = "job_executor"
)

// JobManager handles the job and worker lifecycle.
type JobManager struct {
	// jobRegister maintains a registry of all jobs.
	jobRegistry *job.Registry

	// ctxFactory is used to create new sdk.Context(s).
	ctxFactory *contextFactory

	// Job producers are a pool of workers that produce jobs. These workers
	// run in the background and produce jobs that are then consumed by the
	// job executors.
	producerCfg  *worker.PoolConfig
	jobProducers *worker.Pool

	// Job executors are a pool of workers that execute jobs. These workers
	// are fed jobs by the job producers.
	executorCfg  *worker.PoolConfig
	jobExecutors *worker.Pool

	// TODO: introduce telemetry.Metrics to this struct and the BaseApp.
}

// NewManager creates a new manager.
func NewManager(
	jobs []job.Basic,
	ctxFactory *contextFactory,
) *JobManager {
	m := &JobManager{
		jobRegistry: job.NewRegistry(),
		ctxFactory:  ctxFactory,
	}

	// Register all supplied jobs with the manager.
	for _, j := range jobs {
		if err := m.jobRegistry.Register(j); err != nil {
			panic(err)
		}
	}

	// TODO: read pool configs from the config file.

	// Setup the producer worker pool.
	jobCount := uint16(m.jobRegistry.Count())
	m.producerCfg = &worker.PoolConfig{
		Name:             producerName,
		PrometheusPrefix: producerPromName,
		MinWorkers:       jobCount,
		MaxWorkers:       jobCount + 1,
		ResizingStrategy: producerResizeStrategy,
		MaxQueuedJobs:    jobCount,
	}

	// Setup the executor worker pool.
	m.executorCfg = worker.DefaultPoolConfig()
	m.executorCfg.Name = executorName
	m.executorCfg.PrometheusPrefix = executorPromName

	// Return the manager.
	return m
}

// Logger returns the logger for the baseapp.
func (jm *JobManager) Logger(ctx context.Context) log.Logger {
	return sdk.UnwrapContext(ctx).Logger().With("namespace", "job-manager")
}

// Start calls `Setup` on the jobs in the registry as well as spins up the worker pools.
func (jm *JobManager) Start(ctx context.Context) {
	// We pass in the context in order to handle cancelling the workers. We pass the
	// standard go context and not an sdk.Context here since the context here is just used
	// for cancelling the workers on shutdown.
	logger := jm.ctxFactory.logger
	jm.jobExecutors = worker.NewPool(ctx, logger, jm.executorCfg)
	jm.jobProducers = worker.NewPool(ctx, logger, jm.producerCfg)
}

// Stop calls `Teardown` on the jobs in the registry as well as shut's down all the worker pools.
func (jm *JobManager) Stop() {
	var wg sync.WaitGroup

	// Shutdown producers.
	wg.Add(1)
	go func() {
		defer wg.Done()
		jm.jobProducers.Stop()
		jm.jobProducers = nil
	}()

	// Shutdown executors and call Teardown() if a job has one.
	wg.Add(1)
	go func() {
		defer wg.Done()
		jm.jobExecutors.StopAndWait()
		for _, j := range jm.jobRegistry.Iterate() {
			if tj, ok := j.(job.HasTeardown); ok {
				if err := tj.Teardown(); err != nil {
					panic(err)
				}
			}
		}
		jm.jobExecutors = nil
	}()

	// Wait for both to finish.
	wg.Wait()
}

// RunProducers sets up each job and runs its producer.
func (jm *JobManager) RunProducers(gctx context.Context) {
	ctx := jm.ctxFactory.NewSDKContext(gctx)

	// Load all jobs in registry in the order they were registered.
	orderedJobs, err := jm.jobRegistry.IterateInOrder()
	if err != nil {
		panic(err)
	}

	for _, jobID := range orderedJobs.Keys() {
		j := jm.jobRegistry.Get(jobID)

		// Run the setup for the job if it has one.
		if sj, ok := j.(job.HasSetup); ok {
			if err = sj.Setup(ctx); err != nil {
				panic(err)
			}
		}

		// Submit the job to the job producers based on the job's type. Use retries if the job uses
		// a subscription.
		if wrappedJob := job.WrapJob(j); wrappedJob != nil {
			jm.jobProducers.Submit(jm.producerTask(ctx, wrappedJob))
		} else if subJob, ok := j.(job.Subscribable); ok {
			jm.jobProducers.Submit(jm.withRetry(jm.retryableSubscriber(ctx, subJob)))
		} else if ethSubJob, ok := j.(job.EthSubscribable); ok { //nolint:govet // todo fix.
			jm.jobProducers.Submit(jm.withRetry(jm.retryableEthSubscriber(ctx, ethSubJob)))
		} else if blockHeaderJob, ok := j.(job.BlockHeaderSub); ok { //nolint:govet // todo fix.
			jm.jobProducers.Submit(jm.withRetry(jm.retryableHeaderSubscriber(ctx, blockHeaderJob)))
		} else {
			panic(fmt.Sprintf("unknown job type %s", reflect.TypeOf(j)))
		}
	}
}
