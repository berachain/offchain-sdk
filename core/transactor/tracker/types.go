package tracker

import (
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
