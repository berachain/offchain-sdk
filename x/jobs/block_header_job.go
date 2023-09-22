package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Compile time check to ensure that BlockHeaderWatcher implements job.BlockHeaderSub.
var _ job.BlockHeaderSub = (*BlockHeaderWatcher)(nil)

// BlockHeaderWatcher allows you to subscribe a basic job to a block header event. This job can be
// extended to support application specific logic.
type BlockHeaderWatcher struct {
	sub ethereum.Subscription
}

// NewBlockHeaderWatcher creates a new BlockHeaderWatcher.
func NewBlockHeaderWatcher() *BlockHeaderWatcher {
	return &BlockHeaderWatcher{}
}

// RegistryKey implements job.Basic.
func (*BlockHeaderWatcher) RegistryKey() string {
	return "BlockWatcher"
}

// Execute can be overwritten to introduce custom logic upon receiving new block headers.
//
// Execute implements job.Basic.
func (w *BlockHeaderWatcher) Execute(ctx context.Context, args any) (any, error) {
	newHead := args.(*coretypes.Header)
	if newHead == nil {
		return nil, nil
	}

	sCtx := sdk.UnwrapContext(ctx)
	sCtx.Logger().Info("received new block header", "height", newHead.Number.String())
	return nil, nil
}

// Subscribe implements job.BlockHeaderSub.
func (w *BlockHeaderWatcher) Subscribe(ctx context.Context) (ethereum.Subscription, chan *coretypes.Header) {
	sCtx := sdk.UnwrapContext(ctx)
	headerCh, sub, err := sCtx.Chain().SubscribeNewHead(sCtx)
	if err != nil {
		return nil, nil
	}
	w.sub = sub
	sCtx.Logger().Info("Subscribed to new block headers")
	return sub, headerCh
}

// Unsubscribe implements job.BlockHeaderSub.
func (w *BlockHeaderWatcher) Unsubscribe(context.Context) {
	w.sub.Unsubscribe()
}
