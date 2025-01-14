package worker

import (
	"context"

	"github.com/alitto/pond"
	"github.com/berachain/offchain-sdk/v2/log"
)

// and other functionality to the pool.
type Pool struct {
	name   string
	logger log.Logger
	*pond.WorkerPool
}

type TaskGroup struct {
	*pond.TaskGroupWithContext
}

// NewPool creates a new pool.
func NewPool(ctx context.Context, logger log.Logger, cfg *PoolConfig) *Pool {
	p := &Pool{
		name: cfg.Name,
		WorkerPool: pond.New(
			int(cfg.MaxWorkers),
			int(cfg.MaxQueuedJobs),
			pond.Strategy(resizerFromString(cfg.ResizingStrategy)),
			pond.Context(ctx), // allows for cancelling jobs.
			pond.MinWorkers(int(cfg.MinWorkers)),
			pond.PanicHandler(PanicHandler(logger)),
		),
		logger: logger,
	}
	p.setupMetrics(cfg.PrometheusPrefix)
	return p
}

// Logger returns the logger for the pool.
func (p *Pool) Logger() log.Logger {
	return p.logger.With("namespace", p.name+"-pool")
}

// StopAndWait stops the pool and waits for all workers to finish.
func (p *Pool) StopAndWait() {
	p.Logger().Info("stopping worker pool")
	p.Logger().Info("waiting for workers to finish",
		"running", p.RunningWorkers(), "idle", p.IdleWorkers(),
	)
	defer p.Logger().Info("workers finished")
	p.WorkerPool.StopAndWait()
}

// Stop stops the pool without waiting for all workers to finish. NOTE: Tasks being executed by
// workers will continue until completion (unless the process is terminated). Tasks in the queue
// will not be executed.
func (p *Pool) Stop() {
	p.Logger().Info("stopping worker pool")
	defer p.Logger().Info("workers halted")
	p.WorkerPool.Stop()
}
