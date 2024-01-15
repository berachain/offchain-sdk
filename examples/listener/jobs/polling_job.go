package jobs

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Polling = &Poller{}

// Listener is a simple job that logs the current block when it is run.
type Poller struct {
	Interval time.Duration
}

func (Poller) RegistryKey() string {
	return "Poller"
}

func (w *Poller) IntervalTime(_ context.Context) time.Duration {
	return w.Interval
}

// Execute implements job.Basic.
func (w *Poller) Execute(ctx context.Context, args any) (any, error) {
	sCtx := sdk.UnwrapContext(ctx)

	response, _ := sCtx.Chain().TxPoolContent(ctx)
	pendingTx := response["pending"]
	queuedTx := response["queued"]
	sCtx.Logger().Info("txpool_content", "pending", len(pendingTx), "queued", len(queuedTx))
	if len(pendingTx) > 0 {
		sCtx.Logger().Info("txpool_content", "pending", pendingTx)
		for txID := range pendingTx {
			for _, tx := range pendingTx[txID] {
				sCtx.Logger().Info("txpool_content", "tx", tx.Hash().Hex())
			}
		}
	}

	for txID := range pendingTx {
		for _, tx := range pendingTx[txID] {
			sCtx.DB().Put(tx.Hash().Bytes(), []byte{})
		}
	}

	return nil, nil
}
