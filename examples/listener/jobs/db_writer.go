package jobs

import (
	"context"
	"math/big"

	"github.com/berachain/offchain-sdk/v2/job"
	sdk "github.com/berachain/offchain-sdk/v2/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Basic = &DbWriter{}

// Listener is a simple job that logs the current block when it is run.
type DbWriter struct{}

func (DbWriter) RegistryKey() string {
	return "DBWriter"
}

// Execute implements job.Basic.
func (w *DbWriter) Execute(ctx context.Context, args any) (any, error) {
	sCtx := sdk.UnwrapContext(ctx)
	myBlock, _ := sCtx.Chain().BlockNumber(ctx)
	db := sCtx.DB()
	sCtx.Logger().Info("block", "block", new(big.Int).SetUint64(myBlock).String())
	db.Put([]byte("block"), []byte(new(big.Int).SetUint64(myBlock).String()))
	val, err := db.Get([]byte("block"))
	if err != nil {
		panic(err)
	}
	sCtx.Logger().Info("block read from DB", "block", new(big.Int).SetBytes(val).String())
	return nil, nil
}
