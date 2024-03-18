package tracker

import (
	"context"
	"strings"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Status represents the current status of a tx owned by the transactor.
type Status int

const (
	// These statuses are used while the tx status is still unknown with 100% accuracy.
	StatusQueued   Status = iota - 2
	StatusSending         // The tx is sending (or retrying), equivalent to noncer "acquired".
	StatusInFlight        // The tx has been sent, equivalent to noncer "inFlight".

	// These statuses are used after the transaction status has been confirmed with 100% accuracy.
	StatusPending
	StatusSuccess
	StatusReverted
	StatusStale
)

// Subscriber is an interface that defines methods for handling transaction events.
type Subscriber interface {
	// OnSuccess is called when a transaction has been successfully included in a block.
	OnSuccess(*InFlightTx, *coretypes.Receipt) error
	// OnRevert is called when a transaction has been reverted.
	OnRevert(*InFlightTx, *coretypes.Receipt) error
	// OnStale is called when a transaction becomes stale.
	OnStale(context.Context, *InFlightTx) error
}

// InFlightTx represents a transaction that is currently being tracked by the transactor.
type InFlightTx struct {
	*coretypes.Transaction
	MsgIDs  []string
	Receipt *coretypes.Receipt
	isStale bool
}

// ID returns a unique identifier for the event
func (tx *InFlightTx) ID() string {
	return strings.Join(tx.MsgIDs, " | ")
}

// Status returns the current status of a transaction owned by the transactor.
func (t *InFlightTx) Status() Status {
	if t.Receipt == nil {
		if t.isStale {
			return StatusStale
		}
		return StatusPending
	}

	if t.Receipt.Status == 1 {
		return StatusSuccess
	}

	return StatusReverted
}
