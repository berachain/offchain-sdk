package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Compile time check to ensure that BlockHeaderWatcher implements job.BlockHeaderSub, and
// optionally the basic job's Setup and Teardown methods.
var (
	_ job.BlockHeaderSub = (*BlockHeaderWatcher)(nil)
	_ job.HasSetup       = (*BlockHeaderWatcher)(nil)
	_ job.HasTeardown    = (*BlockHeaderWatcher)(nil)
)

// BlockHeaderWatcher allows you to subscribe a basic job to a block header event.
type BlockHeaderWatcher struct {
	job.Basic
	sub ethereum.Subscription
}

// NewBlockHeaderWatcher creates a new BlockHeaderWatcher.
func NewBlockHeaderWatcher(basic job.Basic) *BlockHeaderWatcher {
	return &BlockHeaderWatcher{
		Basic: basic,
	}
}

func (w *BlockHeaderWatcher) Subscribe(
	ctx context.Context,
) (ethereum.Subscription, chan *coretypes.Header) {
	sCtx := sdk.UnwrapContext(ctx)
	headerCh, sub, err := sCtx.Chain().SubscribeNewHead(sCtx)
	if err != nil {
		return nil, nil
	}
	w.sub = sub

	sCtx.Logger().Info("Subscribed to new block headers")
	return sub, headerCh
}

func (w *BlockHeaderWatcher) Unsubscribe(context.Context) {
	w.sub.Unsubscribe()
}

func (w *BlockHeaderWatcher) Setup(ctx context.Context) error {
	if setupJob, ok := w.Basic.(job.HasSetup); ok {
		return setupJob.Setup(ctx)
	}
	return nil
}

func (w *BlockHeaderWatcher) Teardown() error {
	if setupJob, ok := w.Basic.(job.HasTeardown); ok {
		return setupJob.Teardown()
	}
	return nil
}
