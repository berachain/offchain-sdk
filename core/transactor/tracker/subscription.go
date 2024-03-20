package tracker

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/log"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Subscriber is an interface that defines methods for handling responses from the transactor.
type Subscriber interface {
	// OnError is called when a transaction request fails to build or send.
	OnError(context.Context, *Response) error
	// OnSuccess is called when a transaction has been successfully included in a block.
	OnSuccess(*Response, *coretypes.Receipt) error
	// OnRevert is called when a transaction has been reverted.
	OnRevert(*Response, *coretypes.Receipt) error
	// OnStale is called when a transaction becomes stale.
	OnStale(context.Context, *Response) error
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
func (sub *Subscription) Start(ctx context.Context, ch chan *Response) error {
	// Loop over the channel, handling events as they come in.
	for {
		select {
		case <-ctx.Done():
			// If the context is done, return context error to stop the loop.
			return ctx.Err()
		case e := <-ch:
			// Handle the response based on its Status. // TODO: if there is an error with any of
			// the underlying calls, we should propagate.
			switch e.Status() {
			case StatusError:
				if err := sub.OnError(ctx, e); err != nil {
					sub.logger.Error("failed to handle errored request", "err", err)
				}
			case StatusSuccess:
				// If the transaction was successful, call OnSuccess.
				if err := sub.OnSuccess(e, e.receipt); err != nil {
					sub.logger.Error("failed to handle successful tx", "err", err)
				}
			case StatusReverted:
				// If the transaction was reverted, call OnRevert.
				if err := sub.OnRevert(e, e.receipt); err != nil {
					sub.logger.Error("failed to handle reverted tx", "err", err)
				}
			case StatusStale:
				// If the transaction is stale, call OnStale.
				if err := sub.OnStale(ctx, e); err != nil {
					sub.logger.Error("failed to handle stale tx", "err", err)
				}
			case StatusPending:
				// If the transaction is pending, do nothing.
				time.Sleep(retryPendingBackoff)
			}
		}
	}
}
