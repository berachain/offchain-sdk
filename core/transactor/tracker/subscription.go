package tracker

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/log"
)

// Subscription manages a Subscriber, allowing it to be started and stopped.
type Subscription struct {
	Subscriber
	logger log.Logger
}

func NewSubscription(s Subscriber, logger log.Logger) *Subscription {
	return &Subscription{Subscriber: s, logger: logger}
}

// Start starts the Subscription, listening for transaction events.
func (sub *Subscription) Start(ctx context.Context, ch chan *InFlightTx) error {
	// Loop over the channel, handling events as they come in.
	for {
		var err error
		select {
		case e := <-ch:
			// Handle the event based on its ID.
			switch e.Status() {
			case StatusSuccess:
				// If the transaction was successful, call OnSuccess.
				if err = sub.OnSuccess(e, e.Receipt); err != nil {
					sub.logger.Error("failed to handle successful tx", "err", err)
				}
			case StatusReverted:
				// If the transaction was reverted, call OnRevert.
				if err = sub.OnRevert(e, e.Receipt); err != nil {
					sub.logger.Error("failed to handle reverted tx", "err", err)
				}
			case StatusStale:
				// If the transaction is stale, call OnStale.
				if err = sub.OnStale(ctx, e); err != nil {
					sub.logger.Error("failed to handle stale tx", "err", err)
				}
			case StatusPending:
				// If the transaction is pending, do nothing.
				time.Sleep(retryPendingBackoff)
			}
			// TODO: if there is an error with any of the underlying calls, we should propagate.
		case <-ctx.Done():
			// If the context is done, return context error to stop the loop.
			return ctx.Err()
		}
	}
}
