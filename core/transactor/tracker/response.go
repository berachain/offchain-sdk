package tracker

import (
	"time"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Response represents a transaction that is currently being tracked by the transactor.
type Response struct {
	*coretypes.Transaction

	MsgIDs       []string    // Message IDs that were included in the transaction.
	InitialTimes []time.Time // Times each message was initially fired.
	Error        error       // Build or send error.

	// fields only the tracker will set
	receipt *coretypes.Receipt
	isStale bool
}

// Status returns the current status of a transaction owned by the transactor.
func (t *Response) Status() Status {
	if t.Error != nil {
		return StatusError
	}

	if t.receipt == nil {
		if t.isStale {
			return StatusStale
		}
		return StatusPending
	}

	if t.receipt.Status == 1 {
		return StatusSuccess
	}

	return StatusReverted
}
