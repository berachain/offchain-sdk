package transactor

import (
	"time"

	"github.com/berachain/offchain-sdk/types/queue/sqs"
)

type Config struct {
	// How large an individual batched tx will be (uses multicall contract if > 1).
	TxBatchSize int
	// How long to wait for a batch to be flushed (ideally 1 block time).
	TxBatchTimeout time.Duration
	// Whether we wait the full batch timeout before firing txs. False means we will fire as soon
	// as we reach the desired batch size.
	WaitFullBatchTimeout bool
	// How long to wait to retrieve txs from the queue if it is empty (ideally quick <= 1s).
	EmptyQueueDelay time.Duration
	// what the requireSuccess flag should be set to if using multicall for batching txs.
	MulticallRequireSuccess bool

	// Maximum duration allowed for the tx to be signed (increase this if using a remote signer)
	SignTxTimeout time.Duration

	// How long to wait for the pending nonce (ideally 1 block time).
	PendingNonceInterval time.Duration
	// How long to wait for a tx to hit the mempool (ideally 1-2 block time).
	InMempoolTimeout time.Duration
	// How long to wait for a tx to be mined/confirmed by the chain.
	TxReceiptTimeout time.Duration
	// Whether we should resend txs that are stale (not confirmed after the receipt timeout).
	ResendStaleTxs bool

	// How often to post a snapshot of the transactor system status (ideally 1 block time).
	StatusUpdateInterval time.Duration

	// (Optional) SQS queue config. If left empty, an in-memory queue is used.
	SQS sqs.Config
	// If true, the queue (SQS generates its own) message ID will be used for tracking messages,
	// rather than the optional, user-provided message ID.
	UseQueueMessageID bool
}
