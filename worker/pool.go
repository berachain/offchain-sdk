package worker

import (
	"context"

	"github.com/alitto/pond"
	"github.com/berachain/offchain-sdk/log"
)

// and other functionality to the pool.
type Pool struct {
	name string
	*pond.WorkerPool
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
	}
	p.setupMetrics(cfg.PrometheusPrefix)
	return p
}
