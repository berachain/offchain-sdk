package tracker

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
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
func (r *Response) Status() Status {
	if r.Error != nil {
		return StatusError
	}

	if r.receipt == nil {
		if r.isStale {
			return StatusStale
		}
		return StatusPending
	}

	if r.receipt.Status == 1 {
		return StatusSuccess
	}

	return StatusReverted
}

// Nonce overrides the method on Transaction to avoid dereferencing a nil pointer.
func (r *Response) Nonce() uint64 {
	if r.Transaction != nil {
		return r.Transaction.Nonce()
	}

	return 0
}

// To overrides the method on Transaction to avoid dereferencing a nil pointer.
func (r *Response) To() *common.Address {
	if r.Transaction != nil {
		return r.Transaction.To()
	}

	zeroAddr := common.Address{}
	return &zeroAddr
}

// Hash overrides the method on Transaction to avoid dereferencing a nil pointer.
func (r *Response) Hash() common.Hash {
	if r.Transaction != nil {
		return r.Transaction.Hash()
	}

	return common.Hash{}
}
