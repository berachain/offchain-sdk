package tracker

import (
	"context"
	"sync"
	"time"

	"github.com/berachain/go-utils/utils"
	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/huandu/skiplist"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
)

// noncesCapacity is the capacity of the in-mempool nonces cache.
const noncesCapacity = 10000

// Noncer is a struct that manages nonces for transactions.
type Noncer struct {
	sender    common.Address // The address of the sender.
	ethClient eth.Client     // The Ethereum client.

	// mempool state
	latestPendingNonce uint64
	inMempoolNonces    *lru.Cache[uint64, struct{}]

	// "in-process" nonces
	acquired map[uint64]struct{} // The set of acquired nonces.
	inFlight *skiplist.SkipList  // The list of nonces currently in flight; tx remains in flight
	// until we know from the chain the status of the tx.

	mu              sync.Mutex    // Mutex for thread-safe operations.
	refreshInterval time.Duration // How often to refresh the mempool state.
}

// NewNoncer creates a new Noncer instance.
func NewNoncer(sender common.Address, refreshInterval time.Duration) *Noncer {
	return &Noncer{
		sender:          sender,
		inMempoolNonces: lru.NewCache[uint64, struct{}](noncesCapacity),
		acquired:        make(map[uint64]struct{}),
		inFlight:        skiplist.New(skiplist.Uint64),
		refreshInterval: refreshInterval,
	}
}

func (n *Noncer) Start(ctx context.Context, ethClient eth.Client) {
	n.ethClient = ethClient
	go n.refreshLoop(ctx)
}

func (n *Noncer) refreshLoop(ctx context.Context) {
	n.refreshNonces(ctx)

	ticker := time.NewTicker(n.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.refreshNonces(ctx)
		}
	}
}

// refreshNonces refreshes the pending nonces from the mempool.
func (n *Noncer) refreshNonces(ctx context.Context) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Update the latest pending nonce.
	if pendingNonce, err := n.ethClient.PendingNonceAt(ctx, n.sender); err == nil {
		// This should already be in sync with latest pending nonce according to the chain.
		n.latestPendingNonce = pendingNonce
		// TODO: handle case where stored & chain pending nonce is out of sync?
	}

	// Get all the nonces in the mempool, to notify whether a tx at a given nonce is replacing
	// an existing mempool tx.
	if pendingNonces, err := getPendingNoncesFor(ctx, n.ethClient, n.sender); err == nil {
		for nonce := range pendingNonces {
			n.inMempoolNonces.Add(nonce, struct{}{})
		}
	}
	if queuedNonces, err := getQueuedNoncesFor(ctx, n.ethClient, n.sender); err == nil {
		for nonce := range queuedNonces {
			n.inMempoolNonces.Add(nonce, struct{}{})
		}
	}
}

// Acquire gets the next available nonce. Along with the nonce to use, it returns whether this
// nonce is replacing another tx in the mempool that has the same nonce (in this case, a
// replacement with bumped gas should be used).
func (n *Noncer) Acquire() (uint64, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get the next available nonce from the inFlight list, if any.
	var (
		nonce uint64
		front = n.inFlight.Front()
		back  = n.inFlight.Back()
	)
	if front != nil && back != nil {
		// Iterate through the inFlight objects to ensure there are no gaps
		// TODO: convert to use a binary tree to go from O(n) to O(log(n))
		for nonce = mustNonce(front); nonce <= mustNonce(back); nonce++ {
			if n.inFlight.Get(nonce) == nil {
				// If a gap is found, use that nonce.
				break
			}
		}
	}
	if nonce < n.latestPendingNonce {
		nonce = n.latestPendingNonce
	}
	n.acquired[nonce] = struct{}{}

	// Tx is "replacing" only if the returned nonce is already pending/queued in the mempool.
	return nonce, n.inMempoolNonces.Remove(nonce)
}

// RemoveAcquired removes a nonce from the acquired list, when a transaction is unable to be sent.
func (n *Noncer) RemoveAcquired(nonce uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.acquired, nonce)
}

// SetInFlight adds a transaction to the in-flight list. The transaction is indexed by its nonce.
func (n *Noncer) SetInFlight(nonce uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.acquired, nonce)         // Remove from the acquired nonces.
	n.inFlight.Set(nonce, struct{}{}) // Add to the in-flight list.

	// Update the latest pending nonce.
	n.latestPendingNonce = nonce + 1
}

// RemoveInFlight removes a transaction from the in-flight list by its nonce.
func (n *Noncer) RemoveInFlight(nonce uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inFlight.Remove(nonce)
}

// Stats returns the number of acquired nonces and the number of in-flight transactions.
func (n *Noncer) Stats() (int, int) {
	return len(n.acquired), n.inFlight.Len()
}

// mustNonce returns the nonce of an element from the key.
func mustNonce(element *skiplist.Element) uint64 {
	return utils.MustGetAs[uint64](element.Key())
}
