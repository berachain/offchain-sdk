package tracker

import (
	"context"

	"github.com/berachain/offchain-sdk/log"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Subscriber is an interface that defines methods for handling responses from the transactor.
type Subscriber interface {
	// OnError is called when a transaction request fails to build or send.
	OnError(ctx context.Context, resp *Response)
	// OnSuccess is called when a transaction has been successfully included in a block.
	OnSuccess(resp *Response, receipt *coretypes.Receipt)
	// OnRevert is called when a transaction has been reverted.
	OnRevert(resp *Response, receipt *coretypes.Receipt)
	// OnStale is called when a transaction becomes stale after the configured timeout.
	OnStale(ctx context.Context, resp *Response, isPending bool)
}

// Once started, a Subscription manages and invokes a Subscriber.
type Subscription struct {
	Subscriber
	logger log.Logger
}

func NewSubscription(s Subscriber, logger log.Logger) *Subscription {
	return &Subscription{Subscriber: s, logger: logger}
}

// Start starts the Subscription, listening for transaction events.
//
//nolint:gocognit // okay.
func (sub *Subscription) Start(ctx context.Context, ch chan *Response) {
	// Loop over the channel, handling events as they come in.
	for {
		select {
		case <-ctx.Done():
			// If the context is done, return to stop the loop.
			return
		case e := <-ch:
			// Handle the response based on its Status.
			switch e.Status() {
			case StatusError:
				// If the transaction failed to build or send, call OnError.
				sub.OnError(ctx, e)
			case StatusSuccess:
				// If the transaction is successful, call OnSuccess.
				sub.OnSuccess(e, e.receipt)
			case StatusReverted:
				// If the transaction reverted, call OnRevert.
				sub.OnRevert(e, e.receipt)
			case StatusStale:
				// If the transaction expired from timeout, call OnStale.
				sub.OnStale(ctx, e, false)
			case StatusPending:
				// If the transaction is pending in txPool, call OnStale but with isPending true.
				sub.OnStale(ctx, e, true)
			}
		}
	}
}
