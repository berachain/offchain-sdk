package jobs

import (
	"context"
	"math/big"
	"time"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Polling = &Poller{}

// Listener is a simple job that logs the current block when it is run.
type Poller struct{}

func (w *Poller) IntervalTime(context.Context) time.Duration {
	return 1 * time.Second
}

// Execute implements job.Basic.
func (w *Poller) Execute(ctx context.Context, args any) (any, error) {
	sCtx := sdk.UnwrapSdkContext(ctx)
	myBlock, _ := sCtx.Chain().BlockNumber(ctx)
	sCtx.Logger().Info("block", "block", new(big.Int).SetUint64(myBlock).String())
	return nil, nil
}
