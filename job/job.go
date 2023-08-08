package job

import (
	"context"
)

// Basic represents a basic job. Borrowing the terminology from inheritance, we can
// think of Basic jobs as the Abstract Base Class for all jobs. Basic jobs define execution,
// but do not define any behaviour around how/when the job is to be executed, and thus cannot
// be executed on their own.
type Basic interface {
	RegistryKey() string
	Execute(context.Context, any) (any, error)
}

// HasSetup represents a job that has a setup function.
type HasSetup interface {
	Basic
	Setup(context.Context) error
}

// HasTeardown represents a job that has a teardown function.
type HasTeardown interface {
	Basic
	Teardown() error
}

// HasProducer represents a struct that defines a producer.
type HasProducer interface {
	Basic
	Producer(ctx context.Context, pool WorkerPool) error
}

// HasMetrics represents a struct that defines metrics for
// its internal functions.
type HasMetrics interface {
	Basic
	// RegisterMetrics()
}

// Custom Jobs are jobs that defines their own producer function. This is useful
// for adding custom job types without having to make a change to the core `offchain-sdk`.
type Custom interface {
	Basic
	HasProducer
}
