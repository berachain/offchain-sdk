package tracker

import (
	"context"
	"errors"
	"time"

	"github.com/berachain/offchain-sdk/core/transactor/event"
	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

const retryPendingBackoff = time.Second

type Logger interface{}

// Tracker.
type Tracker struct {
	noncer       *Noncer
	staleTimeout time.Duration
	dispatcher   *event.Dispatcher[*InFlightTx]
}

// NewTracker creates a new transaction tracker.
func New(
	noncer *Noncer, dispatcher *event.Dispatcher[*InFlightTx], staleTimeout time.Duration,
) *Tracker {
	return &Tracker{
		noncer:       noncer,
		staleTimeout: staleTimeout,
		dispatcher:   dispatcher,
	}
}

// AddSubscriber adds a subscriber to the tracker.
func (t *Tracker) Subscribe(ch chan *InFlightTx) {
	t.dispatcher.Subscribe(ch)
}

// Unsubscribe removes a subscriber from the tracker.
func (t *Tracker) Unsubscribe(ch chan *InFlightTx) {
	t.dispatcher.Unsubscribe(ch)
}

// Track adds a transaction to the in-flight list.
func (t *Tracker) Track(
	ctx context.Context, tx *InFlightTx, async bool,
) {
	if async {
		go t.track(ctx, tx)
	} else {
		t.track(ctx, tx)
	}
}

// track adds a transaction to the in-flight list.
func (t *Tracker) track(ctx context.Context, tx *InFlightTx) {
	// If there is already a transaction that is being tracked for this nonce.
	if oldTx := t.noncer.GetInFlight(tx.Nonce()); oldTx != nil {
		// Watch for the old transaction to be replaced.
		if err := t.watchTxForReplacement(ctx, oldTx); err != nil {
			// Need to notify subscribers of this error.
			t.markErr(ctx, tx, err)
			t.dispatcher.Dispatch(tx)
		}
	}

	t.noncer.SetInFlight(tx)
	t.watchTx(ctx, tx)
}

// watchTxForReplacement is watching for a transaction to be replaced by another.
func (t *Tracker) watchTxForReplacement(ctx context.Context, tx *InFlightTx) error {
	sCtx := sdk.UnwrapContext(ctx)
	ethClient := sCtx.Chain()
	// Loop until we see the transaction get replaced.
loop:
	for {
		inMempoolTx, isPending, err := ethClient.TransactionByHash(ctx, tx.Hash())
		switch {
		case inMempoolTx == nil || errors.Is(err, ethereum.NotFound):
			// Desired behaviour: the transaction was replaced.
			// wait for removal from mempool before doing anything
			// make sure that the oldtx gets removed first
			t.noncer.RemoveInFlight(tx)
			break loop
		case isPending:
			// If the transaction is still pending we wait....
			time.Sleep(retryPendingBackoff)
			continue
		case !isPending:
			return errors.New("failed to replace transaction, original tx was included in block")
		}
	}
	return nil
}

func (t *Tracker) watchTx(ctx context.Context, tx *InFlightTx) {
	sCtx := sdk.UnwrapContext(ctx)
	ethClient := sCtx.Chain()
	var (
		receipt *coretypes.Receipt
		err     error
	)

	// We want to notify the dispatcher at the end of this function.
	defer t.dispatcher.Dispatch(tx)

	// Loop until the context is done, the transaction status is determined,
	// or the timeout is reached.
	for {
		select {
		case <-ctx.Done():
			// If the context is done, it could be due to cancellation or other reasons.
			return
		case <-time.After(t.staleTimeout):
			// If the timeout is reached, mark the transaction as stale.
			t.markStale(ctx, tx)
			return
		default:
			// Else check for the receipt again.
			receipt, err = ethClient.TransactionReceipt(ctx, tx.Hash())
			switch {
			case errors.Is(err, ethereum.NotFound):
				time.Sleep(retryPendingBackoff)
				continue
			case err != nil:
				t.markErr(sCtx, tx, err)
			default:
				t.markIncluded(ctx, tx, receipt)
			}
			return
		}
	}
}

// markIncluded is called once a transaction has been included in a block.
func (t *Tracker) markIncluded(
	_ context.Context, tx *InFlightTx, receipt *coretypes.Receipt,
) {
	t.noncer.RemoveInFlight(tx)
	tx.Receipt = receipt
}

// markStale marks a transaction as stale if it's in the in-flight list
// and its nonce is less than the current nonce. It doesn't mark the transaction
// as not in flight, since it's still out in the wild somewhere.
func (t *Tracker) markStale(_ context.Context, tx *InFlightTx) {
	tx.isStale = true
}

// markError notifies the subscriber if there is an error with any of the steps
// in the transaction lifecycle.
func (t *Tracker) markErr(_ context.Context, tx *InFlightTx, err error) {
	tx.err = err
}
