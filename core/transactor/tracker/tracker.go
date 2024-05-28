package tracker

import (
	"context"
	"strconv"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/core/transactor/event"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

const retryBackoff = 500 * time.Millisecond

// Tracker is a component that keeps track of the transactions that are already sent to the chain.
type Tracker struct {
	noncer     *Noncer
	dispatcher *event.Dispatcher[*Response]
	senderAddr common.Address // tx sender address

	inMempoolTimeout time.Duration // for hitting mempool
	staleTimeout     time.Duration // for a tx receipt

	ethClient eth.Client
}

// New creates a new transaction tracker.
func New(
	noncer *Noncer, dispatcher *event.Dispatcher[*Response], sender common.Address,
	inMempoolTimeout, staleTimeout time.Duration,
) *Tracker {
	return &Tracker{
		noncer:           noncer,
		dispatcher:       dispatcher,
		senderAddr:       sender,
		inMempoolTimeout: inMempoolTimeout,
		staleTimeout:     staleTimeout,
	}
}

func (t *Tracker) SetClient(chain eth.Client) {
	t.ethClient = chain
}

// Track adds a transaction response to the in-flight list and waits for a status.
func (t *Tracker) Track(ctx context.Context, resp *Response) {
	t.noncer.SetInFlight(resp.Nonce())
	go t.trackStatus(ctx, resp)
}

// trackStatus polls the for transaction status and updates the in-flight list.
func (t *Tracker) trackStatus(ctx context.Context, resp *Response) {
	var (
		txHash = resp.Hash()
		timer  = time.NewTimer(t.inMempoolTimeout)
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
			// Not found in pending mempool, wait for it to be mined or go stale.
			t.waitMined(ctx, resp, false)
			return
		default:
			// Check in the pending mempool again.
			if t.checkPending(ctx, resp) {
				return
			}

			// Check for the receipt again.
			if receipt, err := t.ethClient.TransactionReceipt(ctx, txHash); err == nil {
				t.markConfirmed(resp, receipt)
				return
			}

			// If not found anywhere, wait for a backoff and try again.
			time.Sleep(retryBackoff)
		}
	}
}

// checkPending marks the tx if its in the pending set in the mempool. Returns true if found.
func (t *Tracker) checkPending(ctx context.Context, resp *Response) bool {
	content, err := t.ethClient.TxPoolContentFrom(ctx, t.senderAddr)
	if err != nil {
		return false
	}

	if senderTxs, ok := content["pending"]; ok {
		if _, isPending := senderTxs[strconv.FormatUint(resp.Nonce(), 10)]; isPending {
			t.markPending(ctx, resp)
			return true
		}
	}

	return false
}

// waitMined waits for a receipt until the transaction is either confirmed or marked stale.
func (t *Tracker) waitMined(ctx context.Context, resp *Response, isAlreadyPending bool) {
	// Check if the tx is in the queued set in the mempool, in which case it can be marked stale.
	if t.checkQueued(ctx, resp) {
		return
	}

	var (
		txHash  = resp.Hash()
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
			// If the timeout is reached, mark the transaction as expired (the tx has been lost and
			// not found anywhere if isAlreadyPending == false).
			t.markExpired(resp, isAlreadyPending)
			return
		default:
			// Else check for the receipt again.
			if receipt, err = t.ethClient.TransactionReceipt(ctx, txHash); err == nil {
				t.markConfirmed(resp, receipt)
				return
			}

			// on any error, search for the receipt after a backoff
			time.Sleep(retryBackoff)
		}
	}
}

// checkQueued marks the tx if its in the queued set in the mempool. Returns true if found.
func (t *Tracker) checkQueued(ctx context.Context, resp *Response) bool {
	content, err := t.ethClient.TxPoolContentFrom(ctx, t.senderAddr)
	if err != nil {
		return false
	}

	if senderTxs, ok := content["queued"]; ok {
		if _, isQueued := senderTxs[strconv.FormatUint(resp.Nonce(), 10)]; isQueued {
			// mark the transaction as expired, but it does exist in the mempool.
			t.markExpired(resp, false)
			return true
		}
	}

	return false
}

// markPending marks the transaction as pending. The transaction is sitting in the "pending" set of
// the mempool --> up to the chain to confirm, remove from inflight.
func (t *Tracker) markPending(ctx context.Context, resp *Response) {
	// Remove from the noncer inFlight set since we know the tx has reached the mempool as
	// executable/pending.
	t.noncer.RemoveInFlight(resp.Nonce())
	t.waitMined(ctx, resp, true)
}

// markConfirmed is called once a transaction has been included in the canonical chain.
func (t *Tracker) markConfirmed(resp *Response, receipt *coretypes.Receipt) {
	// Set the contract address field on the receipt since geth doesn't do this.
	receipt.ContractAddress = *resp.To()
	resp.receipt = receipt
	t.dispatchTx(resp)
}

// markExpired marks a transaction has exceeded the configured timeouts. If pending, it should be
// resent (same tx data, same nonce) with a bumped gas. If stale (i.e. not pending), it should be
// rebuilt (same tx data, new nonce) and resent.
func (t *Tracker) markExpired(resp *Response, isPending bool) {
	resp.isStale = !isPending
	t.dispatchTx(resp)
}

// dispatchTx is called once the tx status is confirmed.
func (t *Tracker) dispatchTx(resp *Response) {
	t.noncer.RemoveInFlight(resp.Nonce())
	t.dispatcher.Dispatch(resp)
}
