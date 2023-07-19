package job

import (
	"context"

	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Basic represents a basic job.
type Basic interface {
	Execute(context.Context, any) (any, error)
}

// Conditional represents a conditional job.
type Conditional interface {
	Basic
	Condition(ctx context.Context) bool
}

// Subscribable represents a subscribable job.
type Subscribable interface {
	Basic
	Subscribe(ctx context.Context) chan any
}

// EthSubscribable represents a subscription to an ethereum event.
type EthSubscribable interface {
	Basic
	Subscribe(ctx context.Context) (ethereum.Subscription, chan coretypes.Log)
	Unsubscribe(ctx context.Context)
}
