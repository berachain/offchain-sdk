package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Basic = &Listener{}

// Listener is a simple job that logs the current block when it is run.
type Listener struct{}

func (w *Listener) Start() error {
	return nil
}

func (w *Listener) Stop() error {
	return nil
}

// Execute implements job.Basic.
func (w *Listener) Execute(ctx context.Context, args any) (any, error) {
	sCtx := sdk.UnwrapSdkContext(ctx)
	myBlock, _ := sCtx.Chain().CurrentBlock()
	sCtx.Logger().Info("block", "block", myBlock.Transactions())
	return nil, nil
}
