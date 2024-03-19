package tracker

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/core/transactor/event"
	"github.com/berachain/offchain-sdk/core/transactor/types"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

const retryPendingBackoff = 500 * time.Millisecond

// Tracker is a component that keeps track of the transactions that are already sent to the chain.
type Tracker struct {
	noncer     *Noncer
	dispatcher *event.Dispatcher[*InFlightTx]

	staleTimeout     time.Duration // for a tx receipt
	inMempoolTimeout time.Duration // for hitting mempool
	ethClient        eth.Client
}

// NewTracker creates a new transaction tracker.
func New(
	noncer *Noncer, dispatcher *event.Dispatcher[*InFlightTx],
	staleTimeout time.Duration, inMempoolTimeout time.Duration,
) *Tracker {
	return &Tracker{
		noncer:           noncer,
		staleTimeout:     staleTimeout,
		inMempoolTimeout: inMempoolTimeout,
		dispatcher:       dispatcher,
	}
}

func (t *Tracker) SetClient(chain eth.Client) {
	t.ethClient = chain
}

// Track adds a transaction to the in-flight list and waits for a status.
func (t *Tracker) Track(ctx context.Context, batch *types.BatchRequest) {
	inFlightTx := &InFlightTx{ // TODO: make the same type.
		Transaction: batch.Transaction, MsgIDs: batch.MsgIDs, TimesFired: batch.TimesFired,
	}
	t.noncer.SetInFlight(inFlightTx)
	go t.trackStatus(ctx, inFlightTx)
}

// trackStatus polls the for transaction status and updates the in-flight list.
func (t *Tracker) trackStatus(ctx context.Context, tx *InFlightTx) {
	var (
		txHash    = tx.Hash()
		txHashHex = txHash.Hex()
		timer     = time.NewTimer(t.inMempoolTimeout)
	)
	defer timer.Stop()

	// Loop until the context is done, the transaction status is determined, or the timeout is
	// reached.
	for {
		select {
		case <-ctx.Done():
			// If the context is done, it could be due to cancellation or other reasons.
			return
		case <-timer.C:
			// Not found in mempool, wait for it to be mined or go stale.
			t.waitMined(ctx, tx, false)
			return
		default:
			// Check the mempool again.
			if content, err := t.ethClient.TxPoolContent(ctx); err == nil {
				if _, isPending := content["pending"][txHashHex]; isPending {
					t.markPending(ctx, tx)
					return
				}

				if _, isQueued := content["queued"][txHashHex]; isQueued {
					// mark the transaction as stale, but it does exist in the mempool.
					t.markStale(tx, false)
					return
				}
			}

			// Check for the receipt again.
			if receipt, err := t.ethClient.TransactionReceipt(ctx, txHash); err == nil {
				t.markConfirmed(tx, receipt)
				return
			}

			// If not found anywhere, wait for a backoff and try again.
			time.Sleep(retryPendingBackoff)
		}
	}
}

// waitMined waits for a receipt until the transaction is either confirmed or marked stale.
func (t *Tracker) waitMined(ctx context.Context, tx *InFlightTx, isAlreadyPending bool) {
	var (
		receipt *coretypes.Receipt
		err     error
		timer   = time.NewTimer(t.staleTimeout)
	)
	defer timer.Stop()

	// Loop until the context is done, the transaction status is determined, or the timeout is
	// reached.
	for {
		select {
		case <-ctx.Done():
			// If the context is done, it could be due to cancellation or other reasons.
			return
		case <-timer.C:
			// If the timeout is reached, mark the transaction as stale (the tx has been lost and
			// not found anywhere if isAlreadyPending == false).
			t.markStale(tx, isAlreadyPending)
			return
		default:
			// Else check for the receipt again.
			if receipt, err = t.ethClient.TransactionReceipt(ctx, tx.Hash()); err == nil {
				t.markConfirmed(tx, receipt)
				return
			}

			// on any error, search for the receipt after a backoff
			time.Sleep(retryPendingBackoff)
		}
	}
}

// markPending marks the transaction as pending. The transaction is sitting in the "pending" set of
// the mempool --> up to the chain to confirm, remove from inflight.
func (t *Tracker) markPending(ctx context.Context, tx *InFlightTx) {
	// Remove from the noncer inFlight set since we know the tx has reached the mempool as
	// executable/pending.
	t.noncer.RemoveInFlight(tx)

	t.waitMined(ctx, tx, true)
}

// markConfirmed is called once a transaction has been included in the canonical chain.
func (t *Tracker) markConfirmed(tx *InFlightTx, receipt *coretypes.Receipt) {
	// Set the contract address field on the receipt since geth doesn't do this.
	if contractAddr := tx.To(); contractAddr != nil && receipt != nil {
		receipt.ContractAddress = *contractAddr
	}

	tx.receipt = receipt
	t.dispatchTx(tx)
}

// markStale marks a stale transaction that needs to be resent if not pending.
func (t *Tracker) markStale(tx *InFlightTx, isPending bool) {
	tx.isStale = !isPending
	t.dispatchTx(tx)
}

// dispatchTx is called once the tx status is confirmed.
func (t *Tracker) dispatchTx(tx *InFlightTx) {
	t.noncer.RemoveInFlight(tx)
	t.dispatcher.Dispatch(tx)
}
