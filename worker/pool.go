package worker

import (
	"github.com/alitto/pond"
	"github.com/berachain/offchain-sdk/log"
)

// PoolConfig is the configuration for a pool.
type PoolConfig struct {
	// Name is the name of the pool.
	Name string
	// MinWorkers is the minimum number of workers that the resizer will
	// shrink the pool down to .
	MinWorkers int
	// MaxWorkers is the maximum number of workers that can be active
	// at the same time.
	MaxWorkers int
	// ResizingStrategy is the methodology used to resize the number of workers
	// in the pool.
	ResizingStrategy string
	// MaxQueuedJobs is the maximum number of jobs that can be queued
	// before the pool starts rejecting jobs.
	MaxQueuedJobs int
}

// DefaultPoolConfig is the default configuration for a pool.
var DefaultPoolConfig = &PoolConfig{
	Name:             "default",
	MinWorkers:       4,  //nolint:gomnd // it's ok.
	MaxWorkers:       32, //nolint:gomnd // it's ok.
	ResizingStrategy: "balanced",
	MaxQueuedJobs:    100, //nolint:gomnd // it's ok.
}

// Pool is a wrapper around a pond.WorkerPool. We use this to add logging
// and other functionality to the pool.
type Pool struct {
	name   string
	logger log.Logger
	*pond.WorkerPool
}

// NewPool creates a new pool.
func NewPool(cfg *PoolConfig, logger log.Logger) *Pool {
	return &Pool{
		name:   cfg.Name,
		logger: logger,
		WorkerPool: pond.New(
			cfg.MaxWorkers,
			cfg.MaxQueuedJobs,
			pond.Strategy(resizerFromString(cfg.ResizingStrategy)),
			pond.MinWorkers(cfg.MinWorkers),
		),
	}
}

// Logger returns the logger for the baseapp.
func (p *Pool) Logger() log.Logger {
	return p.logger.With("namespace", p.name+"-pool")
}

// AddJob adds a job to the pool.
func (p *Pool) AddJob(pay Payload) {
	// We use TrySubmit as to not block the calling thread.
	if !p.TrySubmit(func() { pay.Execute() }) {
		p.Logger().Error("failed to submit job")
	} else {
		p.Logger().Info("submitted job")
	}
}
