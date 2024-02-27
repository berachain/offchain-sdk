package tracker

import (
	"context"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

type Status int

const (
	StatusPending Status = iota
	StatusSuccess
	StatusReverted
	StatusStale
	StatusError
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

type InFlightTx struct {
	*coretypes.Transaction
	MsgIDs  []string
	Receipt *coretypes.Receipt
	err     error
	isStale bool
}

func (t *InFlightTx) String() string {
	return t.Hash().Hex()
}

func (t *InFlightTx) Status() Status {
	if t.err != nil {
		return StatusError
	}

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

func (t *InFlightTx) Err() error {
	return t.err
}
