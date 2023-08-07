package jobs

import (
	"context"
	"math/big"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Basic = &Listener{}

// Listener is a simple job that logs the current block when it is run.
type Listener struct{}

func (w *Listener) RegistryKey() string {
	return "Listener"
}

// Execute implements job.Basic.
func (w *Listener) Execute(ctx context.Context, args any) (any, error) {
	sCtx := sdk.UnwrapSdkContext(ctx)
	myBlock, _ := sCtx.Chain().BlockNumber(ctx)
	sCtx.Logger().Info("block", "block", new(big.Int).SetUint64(myBlock).String())
	return nil, nil
}
