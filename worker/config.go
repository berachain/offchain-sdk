package worker

import "github.com/alitto/pond"

// PoolConfig is the configuration for a pool.
type PoolConfig struct {
	// Name is the name of the pool.
	Name string
	// PrometheusPrefix is the prefix for the prometheus metrics.
	PrometheusPrefix string
	// MinWorkers is the minimum number of workers that the resizer will
	// shrink the pool down to .
	MinWorkers uint16
	// MaxWorkers is the maximum number of workers that can be active
	// at the same time.
	MaxWorkers uint16
	// ResizingStrategy is the methodology used to resize the number of workers
	// in the pool.
	ResizingStrategy string
	// MaxQueuedJobs is the maximum number of jobs that can be queued
	// before the pool starts rejecting jobs.
	MaxQueuedJobs uint16
}

// DefaultPoolConfig is the default configuration for a pool.
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		Name:             "default",
		PrometheusPrefix: "default",
		MinWorkers:       4,  //nolint:gomnd // it's ok.
		MaxWorkers:       32, //nolint:gomnd // it's ok.
		ResizingStrategy: "balanced",
		MaxQueuedJobs:    100, //nolint:gomnd // it's ok.
	}
}

// resizerFromString returns a pond resizer for the given name.
func resizerFromString(name string) pond.ResizingStrategy {
	switch name {
	case "eager":
		return pond.Eager()
	case "lazy":
		return pond.Lazy()
	case "balanced":
		return pond.Balanced()
	default:
		panic("invalid resizer name")
	}
}
