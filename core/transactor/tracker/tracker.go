package tracker

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/core/transactor/event"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

const retryBackoff = 1 * time.Second

// Tracker is a component that keeps track of the transactions that are already sent to the chain.
type Tracker struct {
	noncer     *Noncer
	dispatcher *event.Dispatcher[*Response]
	senderAddr common.Address // tx sender address

	waitingTimeout time.Duration // how long to spin for a tx status

	ethClient eth.Client
}

// New creates a new transaction tracker.
func New(
	noncer *Noncer, dispatcher *event.Dispatcher[*Response], sender common.Address,
	txWaitingTimeout time.Duration,
) *Tracker {
	return &Tracker{
		noncer:         noncer,
		dispatcher:     dispatcher,
		senderAddr:     sender,
		waitingTimeout: txWaitingTimeout,
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

// trackStatus polls the for transaction status (waits for it to reach the mempool or be confirmed)
// and updates the in-flight list.
func (t *Tracker) trackStatus(ctx context.Context, resp *Response) {
	var (
		txHash    = resp.Hash()
		timer     = time.NewTimer(t.waitingTimeout)
		isPending bool
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
			// Not found after waitingTimeout, mark it stale.
			t.markExpired(resp, isPending)
			return
		default:
			// Wait for a backoff before trying again.
			time.Sleep(retryBackoff)

			// Check in the pending mempool, only if we know it's not already pending.
			if !isPending {
				if pendingNonces, err := getPendingNoncesFor(
					ctx, t.ethClient, t.senderAddr,
				); err == nil {
					if _, isPending = pendingNonces[resp.Nonce()]; isPending {
						// Remove from the noncer inFlight set since we know the tx has reached
						// the mempool as executable/pending. Now waiting for confirmation.
						t.noncer.RemoveInFlight(resp.Nonce())
					}
				}
			}

			// Check for the receipt.
			if receipt, err := t.ethClient.TransactionReceipt(ctx, txHash); err == nil {
				t.markConfirmed(resp, receipt)
				return
			}
		}
	}
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
