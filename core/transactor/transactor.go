package transactor

import (
	"context"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/core/transactor/event"
	"github.com/berachain/offchain-sdk/core/transactor/factory"
	"github.com/berachain/offchain-sdk/core/transactor/sender"
	"github.com/berachain/offchain-sdk/core/transactor/tracker"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	"github.com/berachain/offchain-sdk/log"
	sdk "github.com/berachain/offchain-sdk/types"
	kmstypes "github.com/berachain/offchain-sdk/types/kms/types"
	queuetypes "github.com/berachain/offchain-sdk/types/queue/types"

	"github.com/ethereum/go-ethereum/common"
)

// TxrV2 is the main transactor object. TODO: deprecate off being a job.
type TxrV2 struct {
	cfg        Config
	requests   queuetypes.Queue[*types.TxRequest]
	sender     *sender.Sender
	tracker    *tracker.Tracker
	factory    *factory.Factory
	noncer     *tracker.Noncer
	dispatcher *event.Dispatcher[*tracker.InFlightTx]
	logger     log.Logger
	mu         sync.Mutex
}

// NewTransactor creates a new transactor with the given config, request queue, and signer.
func NewTransactor(
	cfg Config, queue queuetypes.Queue[*types.TxRequest], signer kmstypes.TxSigner,
) *TxrV2 {
	noncer := tracker.NewNoncer(signer.Address(), cfg.PendingNonceInterval)
	factory := factory.New(
		noncer, signer,
		factory.NewMulticall3Batcher(common.HexToAddress(cfg.Multicall3Address)),
	)
	dispatcher := event.NewDispatcher[*tracker.InFlightTx]()
	tracker := tracker.New(noncer, dispatcher, cfg.TxReceiptTimeout, cfg.InMempoolTimeout)

	return &TxrV2{
		dispatcher: dispatcher,
		cfg:        cfg,
		factory:    factory,
		sender:     sender.New(factory),
		tracker:    tracker,
		noncer:     noncer,
		requests:   queue,
	}
}

// RegistryKey implements job.Basic.
func (t *TxrV2) RegistryKey() string {
	return "transactor"
}

// Setup implements job.HasSetup.
func (t *TxrV2) Setup(ctx context.Context) error {
	sCtx := sdk.UnwrapContext(ctx)
	chain := sCtx.Chain()
	t.logger = sCtx.Logger()

	// Register the transactor as a subscriber to the tracker.
	ch := make(chan *tracker.InFlightTx)
	go func() {
		subCtx, cancel := context.WithCancel(ctx)
		_ = tracker.NewSubscription(t, t.logger).Start(subCtx, ch) // TODO: handle error
		cancel()
	}()
	t.dispatcher.Subscribe(ch)

	// Setup and start all the transactor components.
	t.factory.SetClient(chain)
	t.sender.Setup(chain, t.logger)
	t.tracker.SetClient(chain)
	t.noncer.Start(ctx, chain)
	go t.mainLoop(ctx)

	return nil
}

// Execute implements job.Basic.
func (t *TxrV2) Execute(_ context.Context, _ any) (any, error) {
	acquired, inFlight := t.noncer.Stats()
	t.logger.Info(
		"🧠 system status",
		"waiting-tx", acquired, "in-flight-tx", inFlight, "pending-requests", t.requests.Len(),
	)
	return nil, nil //nolint:nilnil // its okay.
}

// IntervalTime implements job.Polling.
func (t *TxrV2) IntervalTime(context.Context) time.Duration {
	return t.cfg.StatusUpdateInterval
}

// SubscribeTxResults sends the tx results, once confirmed, to the given subscriber.
func (t *TxrV2) SubscribeTxResults(ctx context.Context, subscriber tracker.Subscriber) {
	ch := make(chan *tracker.InFlightTx)
	go func() {
		subCtx, cancel := context.WithCancel(ctx)
		_ = tracker.NewSubscription(subscriber, t.logger).Start(subCtx, ch) // TODO: handle error
		cancel()
	}()
	t.dispatcher.Subscribe(ch)
}

// SendTxRequest adds the given tx request to the tx queue, after validating it.
func (t *TxrV2) SendTxRequest(txReq *types.TxRequest) (string, error) {
	if err := txReq.Validate(); err != nil {
		return "", err
	}
	return t.requests.Push(txReq)
}

// GetPreconfirmedState returns the status of the given message ID before it has been confirmed by
// the chain. TODO: fix.
func (t *TxrV2) GetPreconfirmedState(msgID string) types.PreconfirmState {
	switch {
	// case t.tracker.IsInFlight(msgID):
	// 	return types.StateInFlight
	// case t.sender.IsSending(msgID):
	// 	return types.StateSending
	// case t.requests.InQueue(msgID):
	// 	return types.StateQueued
	default:
		return types.StateUnknown
	}
}

// Start starts the transactor.
func (t *TxrV2) Start(ctx context.Context) {
	go t.mainLoop(ctx)
}

// mainLoop is the main transaction sending / batching loop.
func (t *TxrV2) mainLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Attempt the retrieve a batch from the queue.
			batch := t.retrieveBatch(ctx)
			if len(batch) == 0 {
				// We didn't get any transactions, so we wait for more.
				t.logger.Info("no tx requests to process....")
				time.Sleep(t.cfg.EmptyQueueDelay)
				continue
			}

			// We got a batch, so we send it and track it. But first wait for the previous sending
			// to finish.
			t.mu.Lock()
			go func() {
				defer t.mu.Unlock()

				// Build the batch request from the factory.
				batchReq, err := t.factory.BuildTransactionFromRequests(ctx, batch...)
				if err != nil {
					t.logger.Error("failed to build batch", "msgs", batch, "err", err)
					return
				}

				// Send and track the batch request.
				if err := t.sendAndTrack(ctx, batchReq); err != nil {
					t.logger.Error("failed to send batch", "msgs", batch, "err", err)
				}
			}()
		}
	}
}

// retrieveBatch retrieves a batch of transaction requests from the queue. It waits until 1) it
// hits the batch timeout or 2) tx batch size is reached only if waitFullBatchTimeout is false.
func (t *TxrV2) retrieveBatch(ctx context.Context) []*types.TxRequest {
	var (
		batch []*types.TxRequest
		timer = time.NewTimer(t.cfg.TxBatchTimeout)
	)
	defer timer.Stop()

	// Loop until the batch tx timeout expires.
	for {
		select {
		case <-ctx.Done():
			return batch
		case <-timer.C:
			return batch
		default:
			txsRemaining := t.cfg.TxBatchSize - len(batch)

			// If we reached max batch size, we can break out of the loop.
			if txsRemaining == 0 {
				// Sleep for the remaining time if we want to wait for the full batch timeout.
				if t.cfg.WaitFullBatchTimeout {
					<-timer.C
				}
				return batch
			}

			// Get at most txsRemaining tx requests from the queue.
			msgIDs, txReqs, err := t.requests.ReceiveMany(int32(txsRemaining))
			if err != nil {
				t.logger.Error("failed to receive tx request", "err", err)
				continue
			}

			// Update the batched tx requests.
			for i, txReq := range txReqs {
				txReq.MsgID = msgIDs[i]
				batch = append(batch, txReq)
			}
		}
	}
}

// sendAndTrack processes a batch of transaction requests. It sends the batch as one transction
// and also tracks the transaction for its status.
func (t *TxrV2) sendAndTrack(ctx context.Context, batch *types.BatchRequest) error {
	// Send the transaction to the chain.
	if err := t.sender.SendTransaction(ctx, batch); err != nil {
		return err
	}

	// Track the transaction status async.
	t.tracker.Track(ctx, batch)

	t.logger.Debug("📡 sent transaction", "hash", batch.Hash().Hex(), "reqs", batch.Len())
	return nil
}
