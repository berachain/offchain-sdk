package tracker

import (
	"context"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/huandu/skiplist"

	"github.com/ethereum/go-ethereum/common"
)

// Noncer is a struct that manages nonces for transactions.
type Noncer struct {
	sender    common.Address     // The address of the sender.
	ethClient eth.Client         // The Ethereum client.
	acquired  *skiplist.SkipList // The list of acquired nonces.
	inFlight  *skiplist.SkipList // The list of nonces currently in flight.
	mu        sync.Mutex         // Mutex for thread-safe operations.

	pendingNonceTimeout  time.Duration
	latestConfirmedNonce uint64
	latestPendingNonce   uint64
}

// NewNoncer creates a new Noncer instance.
func NewNoncer(sender common.Address, pendingNonceTimeout time.Duration) *Noncer {
	return &Noncer{
		sender:              sender,
		acquired:            skiplist.New(skiplist.Uint64),
		inFlight:            skiplist.New(skiplist.Uint64),
		mu:                  sync.Mutex{},
		pendingNonceTimeout: pendingNonceTimeout,
	}
}

func (n *Noncer) SetClient(ethClient eth.Client) {
	n.ethClient = ethClient
}

func (n *Noncer) RefreshLoop(ctx context.Context) {
	n.refreshNonces(ctx)
	ticker := time.NewTicker(n.pendingNonceTimeout)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.refreshNonces(ctx)
		}
	}
}

// refreshNonces refreshes the confirmed and pending nonces.
func (n *Noncer) refreshNonces(ctx context.Context) {
	n.mu.Lock()
	defer n.mu.Unlock()

	confirmedNonce, err := n.ethClient.NonceAt(ctx, n.sender, nil)
	if err == nil {
		n.latestConfirmedNonce = confirmedNonce
	}

	pendingNonce, err := n.ethClient.PendingNonceAt(ctx, n.sender)
	if err == nil {
		n.latestPendingNonce = pendingNonce
	}
}

// Acquire gets the next available nonce.
func (n *Noncer) Acquire(ctx context.Context) (uint64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	var (
		val       = n.inFlight.Back()
		nextNonce uint64
		foundGap  bool
	)
	if val != nil {
		// Iterate through the inFlight objects to ensure there are no gaps
		// TODO: convert to use a binary tree to go from O(n) to O(log(n))
		for i := n.latestConfirmedNonce; i <= val.Value.(*InFlightTx).Nonce(); i++ {
			if n.inFlight.Get(i) == nil {
				// If a gap is found, use that
				nextNonce = i
				foundGap = true
				break
			}
		}
		// If we didn't find a gap, use the next nonce.
		if !foundGap {
			nextNonce = val.Value.(*InFlightTx).Nonce() + 1
		}
	} else {
		nextNonce = n.latestPendingNonce
	}

	n.acquired.Set(nextNonce, nextNonce)
	return nextNonce, nil
}

// SetInFlight adds a transaction to the in-flight list.
// The transaction is indexed by its nonce.
func (n *Noncer) SetInFlight(tx *InFlightTx) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Remove from the acquired nonces.
	n.acquired.Remove(tx.Nonce())

	// Add to the in-flight list.
	n.inFlight.Set(tx.Nonce(), tx)
}

// GetInFlight retrieves a transaction from the in-flight list by its nonce.
// It returns nil if no transaction with the given nonce is found.
func (n *Noncer) GetInFlight(nonce uint64) *InFlightTx {
	n.mu.Lock()
	defer n.mu.Unlock()
	val := n.inFlight.Get(nonce)
	if val == nil {
		return nil
	}
	return val.Value.(*InFlightTx)
}

// InFlight checks if a transaction with the given nonce is in-flight.
// It returns true if the transaction is in-flight, false otherwise.
func (n *Noncer) InFlight(nonce uint64) bool {
	return n.GetInFlight(nonce) != nil
}

// RemoveInFlight removes a transaction from the in-flight list by its nonce.
func (n *Noncer) RemoveInFlight(tx *InFlightTx) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.inFlight.Remove(tx.Nonce())
}

func (n *Noncer) Stats() (int, int) {
	return n.acquired.Len(), n.inFlight.Len()
}
