package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Basic = &BlockWatcher{}

// Listener is a simple job that logs the current block when it is run.
type BlockWatcher struct{}

func (BlockWatcher) RegistryKey() string {
	return "BlockWatcher"
}

// Execute implements job.Basic.
func (w *BlockWatcher) Execute(ctx context.Context, args any) (any, error) {
	newHead := args.(*coretypes.Header)
	if newHead == nil {
		return nil, nil
	}

	sCtx := sdk.UnwrapContext(ctx)
	sCtx.Logger().Info("block", "block", newHead.Number.String())
	return nil, nil
}
