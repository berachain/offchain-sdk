package transactor

import (
	"time"
)

type Config struct {
	// Hex string address of the multicall contract to be used for batched txs.
	Multicall3Address string

	// How large an individual batched tx will be (uses multicall contract if > 1).
	TxBatchSize int
	// How long to wait for a batch to be flushed (ideally 1 block time).
	TxBatchTimeout time.Duration
	// Whether we wait the full batch timeout before firing txs. False means we will fire as soon
	// as we reach the desired batch size.
	WaitFullBatchTimeout bool
	// How long to wait to retrieve txs from the queue if it is empty (ideally quick <= 1s).
	EmptyQueueDelay time.Duration

	// How long to wait for the pending nonce (ideally 1 block time).
	PendingNonceInterval time.Duration
	// How long to wait for a tx to hit the mempool (ideally 1-2 block time).
	InMempoolTimeout time.Duration
	// How long to wait for a tx to be mined/confirmed by the chain.
	TxReceiptTimeout time.Duration

	// How often to post a snapshot of the transactor system status (ideally 1 block time).
	StatusUpdateInterval time.Duration
}
