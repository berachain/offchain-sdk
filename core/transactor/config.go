package transactor

import (
	"time"
)

type Config struct {
	Multicall3Address   string
	TxReceiptTimeout    time.Duration // how long to wait for a tx to be mined
	PendingNonceTimeout time.Duration // how long to wait for the pending nonce
	EmtpyQueueDelay     time.Duration // how long to wait if the queue is empty
	TxBatchSize         int
	TxBatchTimeout      time.Duration // how long to wait for a batch to be flushed
	CallTxTimeout       time.Duration // how long to wait for a eth call result
}
