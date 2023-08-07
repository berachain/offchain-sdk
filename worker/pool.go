package worker

import (
	"context"

	"github.com/alitto/pond"
)

// and other functionality to the pool.
type Pool struct {
	name string
	*pond.WorkerPool
}

// NewPool creates a new pool.
func NewPool(ctx context.Context, cfg *PoolConfig) *Pool {
	p := &Pool{
		name: cfg.Name,
		WorkerPool: pond.New(
			int(cfg.MaxWorkers),
			int(cfg.MaxQueuedJobs),
			pond.Strategy(resizerFromString(cfg.ResizingStrategy)),
			pond.Context(ctx), // allows for cancelling jobs.
			pond.MinWorkers(int(cfg.MinWorkers)),
		),
	}
	p.setupMetrics(cfg.PrometheusPrefix)
	return p
}
