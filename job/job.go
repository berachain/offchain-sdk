package job

import (
	"context"

	"github.com/berachain/offchain-sdk/worker"
)

// Basic represents a basic job.
type Basic interface {
	Setup(context.Context) error
	Teardown() error
	Execute(context.Context, any) (any, error)
}

// Custom Jobs are jobs that defines their own producer function. This is useful
// for adding custom job types without having to make a change to the core `offchain-sdk`.
type Custom interface {
	Basic
	HasProducer
}

// HasPorducer represents a struct that defines a producer.
type HasProducer interface {
	Producer(ctx context.Context, pool worker.Pool) error
}
