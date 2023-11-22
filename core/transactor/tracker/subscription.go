package tracker

import (
	"context"
	"time"

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

// SubscriberWrapper wraps a Subscriber, allowing it to be started and stopped.
type SubscriberWrapper struct {
	Subscriber
}

// Start starts the SubscriberWrapper, listening for transaction events.
func (s *SubscriberWrapper) Start(ctx context.Context, ch chan *InFlightTx) error {
	// Loop over the channel, handling events as they come in.
	for {
		var err error
		select {
		case e := <-ch:
			// Handle the event based on its ID.
			switch e.ID() {
			case int(StatusSuccess):
				// If the transaction was successful, call OnSuccess.
				err = s.OnSuccess(e, e.Receipt)
			case int(StatusReverted):
				// If the transaction was reverted, call OnRevert.
				err = s.OnRevert(e, e.Receipt)
			case int(StatusStale):
				// If the transaction is stale, call OnStale.
				err = s.OnStale(ctx, e)
			case int(StatusError):
				// If there was an error with the transaction, call OnError.
				s.OnError(ctx, e, e.Err())
			case int(StatusPending):
				// If the transaction is pending, do nothing.
				time.Sleep(retryPendingBackoff)
			}
			// TODO: if there is an error with any of the underlying calls, we should handle.
			_ = err
		case <-ctx.Done():
			// If the context is done, return nil to stop the loop.
			return nil
		}
	}
}
