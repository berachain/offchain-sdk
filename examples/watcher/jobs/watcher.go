package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Compile time check to ensure that Watcher implements job.Basic.
var _ job.Basic = &Watcher{}

// Watcher is a simple job that logs the current block when it is run.
type Watcher struct{}

// Execute implements job.Basic.
func (w *Watcher) Execute(ctx context.Context, args any) (any, error) {
	sCtx := sdk.UnwrapSdkContext(ctx)
	myBlock, _ := sCtx.Chain().CurrentBlock()
	sCtx.Logger().Info("block", "block", myBlock.Transactions())
	return nil, nil
}
