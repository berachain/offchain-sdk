package tracker

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/log"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Subscriber is an interface that defines methods for handling transaction events.
type Subscriber interface {
	// OnSuccess is called when a transaction has been successfully included in a block.
	OnSuccess(*InFlightTx, *coretypes.Receipt) error
	// OnRevert is called when a transaction has been reverted.
	OnRevert(*InFlightTx, *coretypes.Receipt) error
	// OnStale is called when a transaction becomes stale.
	OnStale(context.Context, *InFlightTx) error
	// OnError is called when there is an error with the transaction.
	OnError(context.Context, *InFlightTx, error)
}

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
			switch e.ID() {
			case int(StatusSuccess):
				// If the transaction was successful, call OnSuccess.
				if err = sub.OnSuccess(e, e.Receipt); err != nil {
					sub.logger.Error("failed to handle successful tx", "err", err)
				}
			case int(StatusReverted):
				// If the transaction was reverted, call OnRevert.
				if err = sub.OnRevert(e, e.Receipt); err != nil {
					sub.logger.Error("failed to handle reverted tx", "err", err)
				}
			case int(StatusStale):
				// If the transaction is stale, call OnStale.
				if err = sub.OnStale(ctx, e); err != nil {
					sub.logger.Error("failed to handle stale tx", "err", err)
				}
			case int(StatusError):
				// If there was an error with the transaction, call OnError.
				sub.logger.Error("error with transaction", "err", e)
				sub.OnError(ctx, e, e.Err())
			case int(StatusPending):
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
