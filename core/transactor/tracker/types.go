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

func (t *InFlightTx) ID() int {
	if t.err != nil {
		return int(StatusError)
	}

	if t.Receipt == nil {
		if t.isStale {
			return int(StatusStale)
		}
		return int(StatusPending)
	}

	if t.Receipt.Status == 1 {
		return int(StatusSuccess)
	}

	return int(StatusReverted)
}

func (t *InFlightTx) Err() error {
	return t.err
}
