package job

import (
	"github.com/berachain/go-utils/registry"
	"github.com/berachain/go-utils/types"
)

// Registry is a registry of jobtypes to jobs.
type Registry struct {
	// mapRegistry is the underlying map registry of job types to
	types.Registry[string, Basic]
}

// NewRegistry returns a new registry.
func NewRegistry() *Registry {
	return &Registry{
		Registry: registry.NewOrderedMap[string, Basic](),
	}
}

// RegisterJob registers a job type to the registry.
func (r *Registry) RegisterJob(job Basic) {
	if err := r.Registry.Register(job); err != nil {
		panic(err)
	}
}

// Count returns the number of jobs in the registry.
func (r *Registry) Count() uint64 {
	return uint64(len(r.Registry.Iterate()))
}
