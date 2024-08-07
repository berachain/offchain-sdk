package transactor

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/v2/core/transactor/tracker"
	"github.com/berachain/offchain-sdk/v2/core/transactor/types"

	"github.com/ethereum/go-ethereum"
)

// mainLoop is the main transaction sending / batching loop.
func (t *TxrV2) mainLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Attempt the retrieve a batch from the queue.
			requests := t.retrieveBatch(ctx)
			if len(requests) == 0 {
				// We didn't get any transactions, so we wait for more.
				t.logger.Debug("no tx requests to process...")
				time.Sleep(t.cfg.EmptyQueueDelay)
				continue
			}

			// We got a batch, so we can build and fire, after the previous fire has finished.
			go t.fire(
				ctx, &tracker.Response{MsgIDs: requests.MsgIDs(), InitialTimes: requests.Times()},
				true, requests.Messages()...,
			)
		}
	}
}

// retrieveBatch retrieves a batch of transaction requests from the queue. It waits until 1) it
// hits the batch timeout or 2) tx batch size is reached only if waitFullBatchTimeout is false.
func (t *TxrV2) retrieveBatch(ctx context.Context) types.Requests {
	var (
		requests types.Requests
		timer    = time.NewTimer(t.cfg.TxBatchTimeout)
	)
	defer timer.Stop()

	// Loop until the batch tx timeout expires.
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			return requests
		default:
			txsRemaining := t.cfg.TxBatchSize - len(requests)

			// If we reached max batch size, we can break out of the loop.
			if txsRemaining == 0 {
				// Wait until the timer hits if we want to wait for the full batch timeout.
				if t.cfg.WaitFullBatchTimeout {
					<-timer.C
				}
				return requests
			}

			// Get at most txsRemaining tx requests from the queue.
			msgIDs, txReqs, err := t.requests.ReceiveMany(int32(txsRemaining))
			if err != nil {
				t.logger.Error("failed to receive tx request", "err", err)
				continue
			}

			// If using the queue message ID, we need to update the message ID for each tx request.
			if t.cfg.UseQueueMessageID {
				for i, txReq := range txReqs {
					txReq.MsgID = msgIDs[i]
				}
			}

			// Append the tx requests for retrieval.
			requests = append(requests, txReqs...)
		}
	}
}

// fire processes the tracked tx response. If requested to build, it will first batch the messages.
// Then it sends the batch as one tx and asynchronously tracks the tx for its status. Will return
// early and notify tx subscribers if an error occurs during building or sending.
// NOTE: if `toBuild` is false, resp.Transaction must be a valid, signed tx.
// NOTE: this function blocks until any previous calls to `fire` are completed.
func (t *TxrV2) fire(
	ctx context.Context, resp *tracker.Response, toBuild bool, msgs ...*ethereum.CallMsg,
) {
	t.senderMu.Lock()
	defer t.senderMu.Unlock()

	if toBuild {
		// Call the factory to build the (batched) transaction.
		t.markState(types.StateBuilding, resp.MsgIDs...)
		resp.Transaction, resp.Error = t.factory.BuildTransactionFromRequests(ctx, msgs...)
		if resp.Error != nil {
			t.dispatcher.Dispatch(resp)
			return
		}
	}

	// Call the sender to send the transaction to the chain.
	t.markState(types.StateSending, resp.MsgIDs...)
	if resp.Error = t.sender.SendTransaction(ctx, resp.Transaction); resp.Error != nil {
		t.dispatcher.Dispatch(resp)
		return
	}
	t.logger.Info("📡 sent transaction", "hash", resp.Hash().Hex(), "reqs", len(resp.MsgIDs))

	// Call the tracker to track the transaction async.
	t.markState(types.StateInFlight, resp.MsgIDs...)
	t.tracker.Track(ctx, resp)
}
