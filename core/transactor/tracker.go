package transactor

import (
	"context"
	"sync"

	"github.com/berachain/offchain-sdk/core/transactor/tracker"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

func (t *TxrV2) OnSuccess(tx *tracker.InFlightTx, receipt *coretypes.Receipt) error {
	t.logger.Info(
		"⛏️ transaction mined: success",
		"tx-hash", receipt.TxHash.Hex(),
		"gas-used", receipt.GasUsed,
		"status", receipt.Status,
		"block-number", receipt.BlockNumber,
		"nonce", tx.Nonce(),
	)

	// Mark the msgs as processed on the queue in parallel.
	var errs sync.Map
	var wg sync.WaitGroup
	for _, id := range tx.MsgIDs {
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

func (t *TxrV2) OnRevert(tx *tracker.InFlightTx, receipt *coretypes.Receipt) error {
	t.logger.Error(
		"🔻 transaction mined: reverted",
		"tx-hash", receipt.TxHash.Hex(),
		"gas-used", receipt.GasUsed,
		"status", receipt.Status,
		"block-number", receipt.BlockNumber,
		"nonce", tx.Nonce(),
	)

	// The aws queue will move onto the dead queue automatically.
	return nil
}

func (t *TxrV2) OnStale(_ context.Context, tx *tracker.InFlightTx) error {
	t.logger.Warn(
		"🔄 transaction is stale",
		"tx-hash", tx.Hash(),
		"nonce", tx.Nonce(),
		"gas-price", tx.GasPrice(),
	)

	return nil
}

func (t *TxrV2) OnError(_ context.Context, tx *tracker.InFlightTx, _ error) {
	t.logger.Error(
		"❌ error sending transaction",
		"tx-hash", tx.Hash(),
		"nonce", tx.Nonce(),
		"gas-price", tx.GasPrice(),
	)
	// TODO: move ontop dead queue.
}
