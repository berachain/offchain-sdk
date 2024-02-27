package transactor

import (
	"time"
)

type Config struct {
	Multicall3Address   string
	TxReceiptTimeout    time.Duration // how long to wait for a tx to be mined (~2 block time)
	InMempoolTimeout    time.Duration // how long to wait for a tx to hit the mempool (1 block time)
	PendingNonceTimeout time.Duration // how long to wait for the pending nonce (1 block time)
	EmptyQueueDelay     time.Duration // how long to wait if the queue is empty (quick <= 1s)
	TxBatchSize         int
	TxBatchTimeout      time.Duration // how long to wait for a batch to be flushed (1 block time)
	CallTxTimeout       time.Duration // how long to wait for a eth call result
}
