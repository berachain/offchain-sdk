package tracker

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/ethereum/go-ethereum/common"
)

// getPendingNoncesFor returns the nonces that are currently pending in the mempool for the given
// sender.
func getPendingNoncesFor(
	ctx context.Context, ethClient eth.Client, sender common.Address,
) (map[uint64]struct{}, error) {
	contentFrom, err := ethClient.TxPoolContentFrom(ctx, sender)
	if err != nil {
		return nil, err
	}

	pending := make(map[uint64]struct{})
	for nonce := range contentFrom["pending"] {
		pending[nonce] = struct{}{}
	}
	return pending, nil
}

// getQueuedNoncesFor returns the nonces that are currently queued in the mempool for the given
// sender.
func getQueuedNoncesFor(
	ctx context.Context, ethClient eth.Client, sender common.Address,
) (map[uint64]struct{}, error) {
	contentFrom, err := ethClient.TxPoolContentFrom(ctx, sender)
	if err != nil {
		return nil, err
	}

	queued := make(map[uint64]struct{})
	for nonce := range contentFrom["queued"] {
		queued[nonce] = struct{}{}
	}
	return queued, nil
}
