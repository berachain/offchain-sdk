package transactor

import (
	"context"
	"sync"

	"github.com/berachain/offchain-sdk/core/transactor/sender"
	"github.com/berachain/offchain-sdk/core/transactor/tracker"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// OnError is called when a transaction request fails to build or send.
func (t *TxrV2) OnError(_ context.Context, resp *tracker.Response) {
	t.noncer.RemoveAcquired(resp.Nonce())
	t.removeStateTracking(resp.MsgIDs...)
	t.logger.Error("❌ error sending transaction", "err", resp.Error, "msgs", resp.MsgIDs)

	// TODO: move ontop dead queue, for SQS.
}

// OnSuccess is called when a transaction has been successfully included in a block.
func (t *TxrV2) OnSuccess(resp *tracker.Response, receipt *coretypes.Receipt) {
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
}

// OnRevert is called when a transaction has been reverted.
func (t *TxrV2) OnRevert(resp *tracker.Response, receipt *coretypes.Receipt) {
	t.removeStateTracking(resp.MsgIDs...)
	t.logger.Warn(
		"🔻 transaction mined: reverted", "tx-hash", receipt.TxHash.Hex(),
		"gas-used", receipt.GasUsed, "status", receipt.Status, "nonce", resp.Nonce(),
	)

	// TODO: delete from SQS queue / move onto the dead queue?
}

// OnStale is called when a transaction becomes stale after the configured timeout.
func (t *TxrV2) OnStale(ctx context.Context, resp *tracker.Response, isPending bool) {
	t.removeStateTracking(resp.MsgIDs...)
	t.logger.Warn(
		"🔄 transaction is stale", "tx-hash", resp.Hash(),
		"nonce", resp.Nonce(), "gas-price", resp.GasPrice(),
	)

	// For a pending tx "stuck" the mempool, it can only be included by bumping gas.
	if isPending {
		resp.Transaction = sender.BumpGas(resp.Transaction)
	}

	// Try resending the tx to the chain if pending or configured to do so. Rebuild it (same tx
	// data, new nonce) and resend.
	if isPending || t.cfg.ResendStaleTxs {
		go t.fire(ctx, resp, false)
	}
}
