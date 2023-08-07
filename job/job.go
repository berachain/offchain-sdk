package job

import (
	"context"
)

// Basic represents a basic job.
type Basic interface {
	Execute(context.Context, any) (any, error)
}

type HasSetup interface {
	Basic
	Setup(context.Context) error
}

type HasTeardown interface {
	Basic
	Teardown() error
}

// Custom Jobs are jobs that defines their own producer function. This is useful
// for adding custom job types without having to make a change to the core `offchain-sdk`.
type Custom interface {
	Basic
	HasProducer
}

// HasProducer represents a struct that defines a producer.
type HasProducer interface {
	Producer(ctx context.Context, pool WorkerPool) error
}
