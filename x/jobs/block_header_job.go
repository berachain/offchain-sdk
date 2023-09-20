package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Compile time check to ensure that EthEventSub implements job.BlockHeaderSub.
var _ job.BlockHeaderSub = (*BlockHeaderWatcher)(nil)

// BlockHeaderWatcher allows you to subscribe a basic job to a block header event
type BlockHeaderWatcher struct {
	job.Basic
}

// NewBlockHeaderWatcher creates a new BlockHeaderWatcher
func NewBlockHeaderWatcher() *BlockHeaderWatcher {
	return &BlockHeaderWatcher{}
}

func (w *BlockHeaderWatcher) Subscribe(ctx context.Context) (ethereum.Subscription, chan *coretypes.Header) {
	sCtx := sdk.UnwrapContext(ctx)
	headerCh, sub, err := sCtx.Chain().SubscribeNewHead(sCtx)
	if err != nil {
		return nil, nil
	}
	sCtx.Logger().Info("Subscribed to new block headers")
	return sub, headerCh

}

func (w *BlockHeaderWatcher) Unsubscribe(ctx context.Context) {
	// TODO: better way to restart here?
	panic("sub failure")
}
