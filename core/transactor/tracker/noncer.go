package tracker

import (
	"context"
	"sync"

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
}

// NewNoncer creates a new Noncer instance.
func NewNoncer(sender common.Address) *Noncer {
	return &Noncer{
		sender:   sender,
		acquired: skiplist.New(skiplist.Uint64),
		inFlight: skiplist.New(skiplist.Uint64),
		mu:       sync.Mutex{},
	}
}

// Start initiates the nonce synchronization.
func (n *Noncer) SetClient(ethClient eth.Client) {
	n.ethClient = ethClient
}

// Acquire gets the next available nonce.
func (n *Noncer) Acquire(ctx context.Context) (uint64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	val := n.inFlight.Back()

	var nextNonce uint64
	if val != nil {
		nextNonce = val.Value.(*InFlightTx).Nonce() + 1
	} else {
		var err error
		// TODO: doing a network call while holding the lock is a bit dangerous
		nextNonce, err = n.ethClient.PendingNonceAt(ctx, n.sender)
		if err != nil {
			return 0, err
		}
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
