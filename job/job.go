package job

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

type WorkerPool interface {
	AddJob(*Payload)
}

// Basic represents a basic job.
type Basic interface {
	Setup(context.Context) error
	Teardown() error
	Execute(context.Context, any) (any, error)
}

// Subscribable represents a subscribable job.
type Subscribable interface {
	Basic
	Subscribe(ctx context.Context) chan any
}

// Polling represents a polling job.
type Polling interface {
	Basic
	IntervalTime(ctx context.Context) time.Duration
}

// EthSubscribable represents a subscription to an ethereum event.
type EthSubscribable interface {
	Basic
	Subscribe(ctx context.Context) (ethereum.Subscription, chan coretypes.Log)
	Unsubscribe(ctx context.Context)
}
