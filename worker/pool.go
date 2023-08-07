package worker

import (
	"errors"

	"github.com/alitto/pond"
	"github.com/berachain/offchain-sdk/log"
)

// and other functionality to the pool.
type Pool struct {
	name   string
	logger log.Logger
	*pond.WorkerPool
}

// NewPool creates a new pool.
func NewPool(cfg *PoolConfig, logger log.Logger) *Pool {
	p := &Pool{
		name:   cfg.Name,
		logger: logger,
		WorkerPool: pond.New(
			cfg.MaxWorkers,
			cfg.MaxQueuedJobs,
			pond.Strategy(resizerFromString(cfg.ResizingStrategy)),
			pond.MinWorkers(cfg.MinWorkers),
		),
	}
	p.setupMetrics(cfg.PrometheusPrefix)
	return p
}

// Logger returns the logger for the baseapp.
func (p *Pool) Logger() log.Logger {
	return p.logger.With("namespace", p.name+"-pool")
}

// SubmitJob adds a job to the pool.
func (p *Pool) SubmitJob(pay Payload) error {
	// We use TrySubmit as to not block the calling thread.
	if !p.TrySubmit(func() { pay.Execute() }) {
		p.Logger().Error("failed to submit job")
		return errors.New("failed to submit job")
	} else {
		p.Logger().Info("submitted job")
		return nil
	}
}
