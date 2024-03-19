package transactor

import (
	"context"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
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

// TxrV2 is the main transactor object.
type TxrV2 struct {
	cfg        Config
	requests   queuetypes.Queue[*types.TxRequest]
	sender     *sender.Sender
	tracker    *tracker.Tracker
	factory    *factory.Factory
	noncer     *tracker.Noncer
	dispatcher *event.Dispatcher[*tracker.InFlightTx]
	chain      eth.Client
	logger     log.Logger
	mu         sync.Mutex
}

// NewTransactor creates a new transactor with the given config, request queue
// and signer.
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
		sender:     sender.New(factory, tracker),
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
// TODO: deprecate off being a job.
func (t *TxrV2) Setup(ctx context.Context) error {
	sCtx := sdk.UnwrapContext(ctx)
	t.chain = sCtx.Chain()
	t.logger = sCtx.Logger()

	// Register the transactor as a subscriber to the tracker.
	ch := make(chan *tracker.InFlightTx)
	go func() {
		subCtx, cancel := context.WithCancel(ctx)
		_ = tracker.NewSubscription(t, t.logger).Start(subCtx, ch) // TODO: handle error
		cancel()
	}()
	t.dispatcher.Subscribe(ch)

	// TODO: need lock on nonce to support more than one
	t.noncer.SetClient(t.chain)
	t.factory.SetClient(t.chain)
	t.sender.Setup(t.chain, t.logger)
	t.tracker.SetClient(t.chain)
	t.Start(sCtx)
	return nil
}

// Execute implements job.Basic.
// TODO: deprecate off being a job.
func (t *TxrV2) Execute(_ context.Context, _ any) (any, error) {
	acquired, inFlight := t.noncer.Stats()
	t.logger.Info(
		"🧠 system status", "waiting-tx", acquired, "in-flight-tx",
		inFlight, "pending-requests", t.requests.Len(),
	)
	return nil, nil //nolint:nilnil // its okay.
}

// IntervalTime implements job.Polling.
func (t *TxrV2) IntervalTime(_ context.Context) time.Duration {
	return 5 * time.Second //nolint:gomnd // TODO: read from config.
}

// SubscribeTxResults sends the tx results (inflight) to the given channel.
func (t *TxrV2) SubscribeTxResults(ctx context.Context, subscriber tracker.Subscriber) {
	ch := make(chan *tracker.InFlightTx)
	go func() {
		subCtx, cancel := context.WithCancel(ctx)
		_ = tracker.NewSubscription(subscriber, t.logger).Start(subCtx, ch) // TODO: handle error
		cancel()
	}()
	t.dispatcher.Subscribe(ch)
}

// SendTxRequest adds the given tx request to the tx queue.
func (t *TxrV2) SendTxRequest(txReq *types.TxRequest) (string, error) {
	return t.requests.Push(txReq)
}

// GetPreconfirmedState returns the status of the given message ID before it has been confirmed by
// the chain.
func (t *TxrV2) GetPreconfirmedState(msgID string) tracker.PreconfirmState {
	switch {
	case t.tracker.IsInFlight(msgID):
		return tracker.StateInFlight
	case t.sender.IsSending(msgID):
		return tracker.StateSending
	case t.requests.InQueue(msgID):
		return tracker.StateQueued
	default:
		return tracker.StateUnknown
	}
}

// Start starts the transactor.
func (t *TxrV2) Start(ctx context.Context) {
	go t.noncer.RefreshLoop(ctx)
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
			msgIDs, timesFired, batch := t.retrieveBatch()

			// We didn't get any transactions, so we wait for more.
			if len(batch) == 0 {
				t.logger.Info("no tx requests to process....")
				time.Sleep(t.cfg.EmptyQueueDelay)
				continue
			}

			// We got a batch, so we send it and track it.
			// We must first wait for the previous sending to finish.
			t.mu.Lock()
			go func() {
				defer t.mu.Unlock()
				if err := t.sendAndTrack(ctx, msgIDs, timesFired, batch...); err != nil {
					t.logger.Error("failed to process batch", "msgs", msgIDs, "err", err)
				}
			}()
		}
	}
}

// retrieveBatch retrieves a batch of transaction requests from the queue. It waits until 1) it
// hits the batch timeout or 2) max batch size only if waitBatchTimeout is false.
func (t *TxrV2) retrieveBatch() ([]string, []time.Time, []*types.TxRequest) {
	var (
		retMsgIDs     []string
		timesFired    []time.Time
		batch         []*types.TxRequest
		startTime     = time.Now()
		timeRemaining = t.cfg.TxBatchTimeout - time.Since(startTime)
	)

	// Loop until the batch tx timeout expires.
	for ; timeRemaining > 0; timeRemaining = t.cfg.TxBatchTimeout - time.Since(startTime) {
		txsRemaining := int32(t.cfg.TxBatchSize - len(batch))
		if txsRemaining == 0 { // if we reached max batch size, we can break out of the loop.
			if t.cfg.WaitBatchTimeout {
				time.Sleep(timeRemaining)
			}
			break
		}

		msgIDs, txReq, times, err := t.requests.ReceiveMany(txsRemaining)
		if err != nil {
			t.logger.Error("failed to receive tx request", "err", err)
			continue
		}

		retMsgIDs = append(retMsgIDs, msgIDs...)
		timesFired = append(timesFired, times...)
		batch = append(batch, txReq...)
	}

	return retMsgIDs, timesFired, batch
}

// sendAndTrack processes a batch of transaction requests.
// It builds a transaction from the batch and sends it.
// It also tracks the transaction for future reference.
func (t *TxrV2) sendAndTrack(
	ctx context.Context, msgIDs []string, timesFired []time.Time, batch ...*types.TxRequest,
) error {
	tx, err := t.factory.BuildTransactionFromRequests(ctx, 0, batch...)
	if err != nil {
		return err
	}

	// Send the transaction to the chain and track it async.
	if err = t.sender.SendTransactionAndTrack(ctx, tx, msgIDs, timesFired, true); err != nil {
		return err
	}

	t.logger.Debug("📡 sent transaction", "tx-hash", tx.Hash().Hex(), "tx-reqs", len(batch))
	return nil
}
