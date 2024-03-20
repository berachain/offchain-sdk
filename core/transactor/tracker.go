package transactor

import (
	"context"
	"sync"

	"github.com/berachain/offchain-sdk/core/transactor/tracker"
	"github.com/berachain/offchain-sdk/core/transactor/types"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

func (t *TxrV2) OnError(_ context.Context, resp *tracker.Response) error {
	t.removeStateTracking(resp.MsgIDs...)
	t.logger.Error("❌ error sending transaction", "err", resp.Error, "msgs", resp.MsgIDs)

	// TODO: move ontop dead queue.
	return nil
}

func (t *TxrV2) OnSuccess(resp *tracker.Response, receipt *coretypes.Receipt) error {
	t.removeStateTracking(resp.MsgIDs...)
	t.logger.Info(
		"⛏️ transaction mined: success", "tx-hash", receipt.TxHash.Hex(),
		"gas-used", receipt.GasUsed, "status", receipt.Status, "nonce", resp.Nonce(),
	)

	// Mark the msgs as processed on the queue in parallel.
	var errs sync.Map
	var wg sync.WaitGroup
	for _, id := range resp.MsgIDs {
		wg.Add(1)
		go func(_id string) {
			defer wg.Done()
			if err := t.requests.Delete(_id); err != nil {
				errs.Store(_id, err)
			}
		}(id)
	}
	wg.Wait()

	// Log any errors that occurred during deletion.
	errs.Range(func(key, value interface{}) bool {
		t.logger.Error("error deleting request from queue", "id", key, "err", value)
		return true
	})
	return nil
}

func (t *TxrV2) OnRevert(resp *tracker.Response, receipt *coretypes.Receipt) error {
	t.removeStateTracking(resp.MsgIDs...)
	t.logger.Error(
		"🔻 transaction mined: reverted", "tx-hash", receipt.TxHash.Hex(),
		"gas-used", receipt.GasUsed, "status", receipt.Status, "nonce", resp.Nonce(),
	)

	// TODO: delete from sqs queue / move onto the dead queue?
	return nil
}

func (t *TxrV2) OnStale(ctx context.Context, resp *tracker.Response) error {
	t.removeStateTracking(resp.MsgIDs...)
	t.logger.Warn(
		"🔄 transaction is stale", "tx-hash", resp.Hash(),
		"nonce", resp.Nonce(), "gas-price", resp.GasPrice(),
	)

	// Try resending the tx to the chain if configured to do so.
	if t.cfg.ResendStaleTxs {
		t.fire(ctx, resp, types.CallMsgFromTx(resp.Transaction))
	}
	return nil
}
