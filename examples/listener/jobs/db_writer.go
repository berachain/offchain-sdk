package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Basic = &DbWriter{}

// Listener is a simple job that logs the current block when it is run.
type DbWriter struct{}

func (w *DbWriter) Start() error {
	return nil
}

func (w *DbWriter) Stop() error {
	return nil
}

// Execute implements job.Basic.
func (w *DbWriter) Execute(ctx context.Context, args any) (any, error) {
	sCtx := sdk.UnwrapSdkContext(ctx)
	myBlock, _ := sCtx.Chain().CurrentBlock()
	db := sCtx.DB()
	sCtx.Logger().Info("block", "block", myBlock.Transactions())
	db.Put([]byte("block"), []byte(myBlock.Number().String()))
	val, err := db.Get([]byte("block"))
	if err != nil {
		panic(err)
	}
	sCtx.Logger().Info("block read from DB", "block", string(val))
	return nil, nil
}
