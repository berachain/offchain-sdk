package tracker

import (
	"context"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Status represents the current status of a tx owned by the transactor. These are used only after
// the tx status has been confirmed by erroring, the chain, or the configured timeout.
type Status uint8

const (
	StatusPending Status = iota
	StatusError
	StatusSuccess
	StatusReverted
	StatusStale
)

// Subscriber is an interface that defines methods for handling responses from the transactor.
type Subscriber interface {
	// OnError is called when a transaction request fails to build or send.
	OnError(ctx context.Context, resp *Response) error
	// OnSuccess is called when a transaction has been successfully included in a block.
	OnSuccess(resp *Response, receipt *coretypes.Receipt) error
	// OnRevert is called when a transaction has been reverted.
	OnRevert(resp *Response, receipt *coretypes.Receipt) error
	// OnStale is called when a transaction becomes stale after the configured timeout.
	OnStale(ctx context.Context, resp *Response, isPending bool) error
}
