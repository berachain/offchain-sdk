package factory

import (
	"context"

	"github.com/berachain/offchain-sdk/core/transactor/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// Noncer is an interface for acquiring fresh nonces.
type Noncer interface {
	Acquire() (uint64, bool)
}

// Batcher is an interface for batching requests, commonly implemented by multicallers.
type Batcher interface {
	// BatchRequests creates a batched transaction request for the given call requests.
	BatchRequests(callReqs ...*ethereum.CallMsg) *types.Request

	// BatchCallRequests returns multicall results after executing the given call requests.
	BatchCallRequests(
		ctx context.Context, from common.Address, callReqs ...*ethereum.CallMsg,
	) (any, error)
}
