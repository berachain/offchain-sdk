package baseapp

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/job"
	workertypes "github.com/berachain/offchain-sdk/job/types"
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

	maxBackoff   = 2 * time.Minute
	backoffStart = 1 * time.Second
	backoffBase  = 2
	jitterRange  = 1000
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

// Start calls `Setup` on the jobs in the registry as well as spins up
// the worker pools.
func (jm *JobManager) Start(ctx context.Context) {
	// We pass in the context in order to handle cancelling the workers. We pass the
	// standard go context and not an sdk.Context here since the context here is just used
	// for cancelling the workers on shutdown.
	logger := jm.ctxFactory.logger
	jm.jobExecutors = worker.NewPool(ctx, logger, jm.executorCfg)
	jm.jobProducers = worker.NewPool(ctx, logger, jm.producerCfg)
}

// Stop calls `Teardown` on the jobs in the registry as well as
// shut's down all the worker pools.
func (jm *JobManager) Stop() {
	var wg sync.WaitGroup

	// Shutdown producers.
	wg.Add(1)
	go func() {
		defer wg.Done()
		jm.jobProducers.Stop()
		jm.jobProducers = nil
	}()

	// Shutdown executors and call Teardown().
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

	// Wait for both to finish. (Should we add a timeout?)
	wg.Wait()
}

func (jm *JobManager) runProducer(ctx context.Context, j job.Basic) bool {
	// Handle migrated jobs.
	if wrappedJob := job.WrapJob(j); wrappedJob != nil {
		jm.jobProducers.Submit(
			func() {
				if err := wrappedJob.Producer(
					ctx, jm.jobExecutors,
				); !errors.Is(err, context.Canceled) && err != nil {
					jm.Logger(ctx).Error("error in job producer", "err", err)
				}
			},
		)
		return true
	}
	return false
}

// RunProducers sets up each job and runs its producer.
func (jm *JobManager) RunProducers(gctx context.Context) {
	for _, j := range jm.jobRegistry.Iterate() {
		ctx := jm.ctxFactory.NewSDKContext(gctx)
		if sj, ok := j.(job.HasSetup); ok {
			if err := sj.Setup(ctx); err != nil {
				panic(err)
			}
		}

		if jm.runProducer(ctx, j) {
			continue
		} else if subJob, ok := j.(job.Subscribable); ok {
			jm.jobProducers.Submit(func() {
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
		} else if ethSubJob, okEthSubJob := j.(job.EthSubscribable); okEthSubJob {
			jm.runEthSubscribable(ctx, ethSubJob)
		} else {
			panic(fmt.Sprintf("unknown job type %s", reflect.TypeOf(j)))
		}
	}
}

func (jm *JobManager) runEthSubscribable(ctx context.Context, j job.EthSubscribable) {
	jm.jobProducers.Submit(withRetry(func() bool {
		sub, ch, err := j.Subscribe(ctx)
		if err != nil {
			jm.Logger(ctx).Error("error in eth subscription", "err", err)
			return true
		}

		for {
			select {
			case <-ctx.Done():
				j.Unsubscribe(ctx)
				return false
			case err = <-sub.Err():
				jm.Logger(ctx).Error("error in eth subscription", "err", err)
				j.Unsubscribe(ctx)
				return true
			case val := <-ch:
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, j, val).Execute)
				continue
			}
		}
	}, jm.Logger(ctx)))
}

// withRetry is a wrapper that retries a task with exponential backoff.
func withRetry(task func() bool, logger log.Logger) func() {
	return func() {
		backoff := backoffStart

		for {
			if retry := task(); retry {
				// Exponential backoff with jitter.
				jitter, _ := rand.Int(rand.Reader, big.NewInt(jitterRange))
				sleep := backoff + time.Duration(jitter.Int64())*time.Millisecond
				logger.Info(fmt.Sprintf("retrying task in %s...", sleep))
				time.Sleep(sleep)
				backoff *= backoffBase
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
			break
		}
	}
}
