package transactor

import (
	"context"
	"sync"

	"github.com/berachain/offchain-sdk/core/transactor/tracker"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	sdk "github.com/berachain/offchain-sdk/types"

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

func (t *TxrV2) OnStale(
	ctx context.Context, inFlightTx *tracker.InFlightTx,
) error {
	t.logger.Warn(
		"🔄 transaction is stale", "tx-hash", inFlightTx.Hash(),
		"nonce", inFlightTx.Nonce(), "gas-price", inFlightTx.GasPrice(),
	)

	var (
		sCtx = sdk.NewContext(ctx, t.chain, t.logger, nil)
		tx   *coretypes.Transaction
		err  error
	)


	tx, err = t.factory.BuildTransactionFromRequests(sCtx, &types.TxRequest{
		To:        inFlightTx.To(),
		Value:     inFlightTx.Value(),
		Data:      inFlightTx.Data(),
		Gas:       inFlightTx.Gas(),
		GasFeeCap: inFlightTx.GasFeeCap(),
		GasTipCap: inFlightTx.GasTipCap(),
		GasPrice:  inFlightTx.GasPrice(),
	})
	if err == nil {
		return t.sender.SendTransactionAndTrack(sCtx, tx, inFlightTx.MsgIDs, true)
	}
	return err
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
