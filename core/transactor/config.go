package transactor

import (
	"time"
)

type Config struct {
	Multicall3Address string

	// How large an individual batched tx will be (uses multicall if > 1).
	TxBatchSize int
	// How long to wait for a batch to be flushed (ideally 1 block time).
	TxBatchTimeout time.Duration
	// How long to wait if the queue is empty (ideally quick <= 1s).
	EmptyQueueDelay time.Duration
	// Whether we wait the full batch timeout before firing txs. False means we will fire as soon
	// as we reach the desired batch size.
	WaitBatchTimeout bool

	// How long to wait for the pending nonce (ideally 1 block time).
	PendingNonceInterval time.Duration
	// How long to wait for a tx to hit the mempool (ideally 1-2 block time).
	InMempoolTimeout time.Duration
	// How long to wait for a tx to be mined/confirmed by the chain.
	TxReceiptTimeout time.Duration
}
