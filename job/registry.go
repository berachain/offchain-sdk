package job

import (
	"github.com/berachain/go-utils/registry"
)

// TODO define this interface in go-utils.
type mapRegistry interface {
	Get(name string) Basic
	Register(Basic) error
	Iterate() map[string]Basic
}

// Registry is a registry of jobtypes to jobs.
type Registry struct {
	// mapRegistry is the underlying map registry of job types to
	mapRegistry
}

// NewRegistry returns a new registry.
func NewRegistry() *Registry {
	return &Registry{
		mapRegistry: registry.NewMap[string, Basic](),
	}
}

// RegisterJob registers a job type to the registry.
func (r *Registry) RegisterJob(job Basic) {
	if err := r.mapRegistry.Register(job); err != nil {
		panic(err)
	}
}

// Count returns the number of jobs in the registry.
func (r *Registry) Count() uint64 {
	return uint64(len(r.mapRegistry.Iterate()))
}
