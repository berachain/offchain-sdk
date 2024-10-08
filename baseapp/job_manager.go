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
	producerResizeStrategy  = "eager"
	executorName           = "job-executor"
	executorPromName       = "job_executor"
)

// JobManager handles the job and worker lifecycle.
type JobManager struct {
	jobRegistry *job.Registry
	ctxFactory  *contextFactory
	producerCfg *worker.PoolConfig
	jobProducers *worker.Pool
	executorCfg *worker.PoolConfig
	jobExecutors *worker.Pool
}

// NewManager creates a new manager.
func NewManager(jobs []job.Basic, ctxFactory *contextFactory) *JobManager {
	m := &JobManager{
		jobRegistry: job.NewRegistry(),
		ctxFactory:  ctxFactory,
	}

	for _, j := range jobs {
		if err := m.jobRegistry.Register(j); err != nil {
			panic(err)
		}
	}

	m.setupProducerPool()
	m.setupExecutorPool()

	return m
}

// setupProducerPool configures the producer worker pool.
func (jm *JobManager) setupProducerPool() {
	jobCount := uint16(jm.jobRegistry.Count()) //nolint:gosec // safe to convert.
	jm.producerCfg = &worker.PoolConfig{
		Name:             producerName,
		PrometheusPrefix: producerPromName,
		MinWorkers:       jobCount,
		MaxWorkers:       jobCount + 1,
		ResizingStrategy: producerResizeStrategy,
		MaxQueuedJobs:    jobCount,
	}
}

// setupExecutorPool configures the executor worker pool.
func (jm *JobManager) setupExecutorPool() {
	jm.executorCfg = worker.DefaultPoolConfig()
	jm.executorCfg.Name = executorName
	jm.executorCfg.PrometheusPrefix = executorPromName
}

// Logger returns the logger for the baseapp.
func (jm *JobManager) Logger(ctx context.Context) log.Logger {
	return sdk.UnwrapContext(ctx).Logger().With("namespace", "job-manager")
}

// Start initializes the job manager and starts the worker pools.
func (jm *JobManager) Start(ctx context.Context) {
	logger := jm.ctxFactory.logger
	jm.jobExecutors = worker.NewPool(ctx, logger, jm.executorCfg)
	jm.jobProducers = worker.NewPool(ctx, logger, jm.producerCfg)
}

// Stop shuts down the worker pools and calls Teardown on jobs in the registry.
func (jm *JobManager) Stop() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		jm.jobProducers.Stop()
		jm.jobProducers = nil
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		jm.jobExecutors.StopAndWait()
		jm.teardownJobs()
		jm.jobExecutors = nil
	}()

	wg.Wait()

	if err := jm.ctxFactory.metrics.Close(); err != nil {
		jm.ctxFactory.logger.Error("failed to close metrics", "err", err)
	}
}

// teardownJobs calls Teardown on jobs in the registry.
func (jm *JobManager) teardownJobs() {
	for _, j := range jm.jobRegistry.Iterate() {
		if tj, ok := j.(job.HasTeardown); ok {
			if err := tj.Teardown(); err != nil {
				panic(err)
			}
		}
	}
}

// RunProducers sets up each job and runs its producer.
func (jm *JobManager) RunProducers(gctx context.Context) {
	ctx := jm.ctxFactory.NewSDKContext(gctx)
	orderedJobs, err := jm.jobRegistry.IterateInOrder()
	if err != nil {
		panic(err)
	}

	for _, jobID := range orderedJobs.Keys() {
		j := jm.jobRegistry.Get(jobID)

		if sj, ok := j.(job.HasSetup); ok {
			if err = sj.Setup(ctx); err != nil {
				panic(err)
			}
		}

		wrappedJob := job.WrapJob(j)
		switch {
		case wrappedJob != nil:
			jm.jobProducers.Submit(jm.producerTask(ctx, wrappedJob))
		case subJob, ok := j.(job.Subscribable); ok:
			jm.jobProducers.Submit(jm.withRetry(jm.retryableSubscriber(ctx, subJob)))
		case ethSubJob, ok := j.(job.EthSubscribable); ok:
			jm.jobProducers.Submit(jm.withRetry(jm.retryableEthSubscriber(ctx, ethSubJob)))
		case blockHeaderJob, ok := j.(job.BlockHeaderSub); ok:
			jm.jobProducers.Submit(jm.withRetry(jm.retryableHeaderSubscriber(ctx, blockHeaderJob)))
		default:
			panic(fmt.Sprintf("unknown job type %s", reflect.TypeOf(j)))
		}
	}
}
