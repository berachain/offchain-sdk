package transactor

import (
	"context"
	"errors"
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
	"github.com/berachain/offchain-sdk/types/queue/mem"
	"github.com/berachain/offchain-sdk/types/queue/sqs"
	queuetypes "github.com/berachain/offchain-sdk/types/queue/types"
	"github.com/ethereum/go-ethereum/common"
)

// TxrV2 is the main transactor object. TODO: deprecate off being a job.
type TxrV2 struct {
	cfg        Config
	logger     log.Logger
	signerAddr common.Address

	requests   queuetypes.Queue[*types.Request]
	factory    *factory.Factory
	noncer     *tracker.Noncer
	sender     *sender.Sender
	senderMu   sync.Mutex
	dispatcher *event.Dispatcher[*tracker.Response]
	tracker    *tracker.Tracker

	preconfirmedStates map[string]types.PreconfirmedState
	preconfirmedMu     sync.RWMutex
}

// NewTransactor creates a new transactor with the given config and signer.
func NewTransactor(cfg Config, signer kmstypes.TxSigner, batcher factory.Batcher) (*TxrV2, error) {
	// Determine queue type based on given configuration.
	var queue queuetypes.Queue[*types.Request]
	if cfg.SQS.QueueURL != "" {
		var err error
		if queue, err = sqs.NewQueueFromConfig[*types.Request](cfg.SQS); err != nil {
			return nil, err
		}
	} else {
		queue = mem.NewQueue[*types.Request]()
	}

	// Ensure a batcher is provided if batching is required.
	if cfg.TxBatchSize > 1 && batcher == nil {
		return nil, errors.New("batcher must be provided when tx batch size is greater than 1")
	}

	// Build the transactor components.
	noncer := tracker.NewNoncer(signer.Address(), cfg.PendingNonceInterval)
	factory := factory.New(noncer, batcher, signer, cfg.SignTxTimeout)
	dispatcher := event.NewDispatcher[*tracker.Response]()
	tracker := tracker.New(
		noncer, dispatcher, signer.Address(), cfg.InMempoolTimeout, cfg.TxReceiptTimeout,
	)

	return &TxrV2{
		cfg:                cfg,
		requests:           queue,
		signerAddr:         signer.Address(),
		factory:            factory,
		noncer:             noncer,
		sender:             sender.New(factory, noncer),
		dispatcher:         dispatcher,
		tracker:            tracker,
		preconfirmedStates: make(map[string]types.PreconfirmedState),
	}, nil
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
	ch := make(chan *tracker.Response)
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

	// If there are any pending txns at startup, they are likely to be stuck in the mempool.
	// Resend them.
	if err := t.resendStaleTxns(ctx); err != nil {
		return err
	}

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
	ch := make(chan *tracker.Response)
	go func() {
		subCtx, cancel := context.WithCancel(ctx)
		_ = tracker.NewSubscription(subscriber, t.logger).Start(subCtx, ch) // TODO: handle error
		cancel()
	}()
	t.dispatcher.Subscribe(ch)
}

// SendTxRequest adds the given tx request to the tx queue, after validating it.
func (t *TxrV2) SendTxRequest(txReq *types.Request) (string, error) {
	if err := txReq.Validate(); err != nil {
		return "", err
	}

	msgID := txReq.MsgID
	queueID, err := t.requests.Push(txReq)
	if err != nil {
		return "", err
	}
	if t.cfg.UseQueueMessageID {
		msgID = queueID
	}

	t.markState(types.StateQueued, msgID)
	return msgID, nil
}

// GetPreconfirmedState returns the status of the given message ID before it has been confirmed by
// the chain.
func (t *TxrV2) GetPreconfirmedState(msgID string) types.PreconfirmedState {
	t.preconfirmedMu.RLock()
	defer t.preconfirmedMu.RUnlock()

	return t.preconfirmedStates[msgID]
}

// markState marks the given preconfirmed state for the given message IDs.
func (t *TxrV2) markState(state types.PreconfirmedState, msgIDs ...string) {
	t.preconfirmedMu.Lock()
	defer t.preconfirmedMu.Unlock()

	for _, msgID := range msgIDs {
		t.preconfirmedStates[msgID] = state
	}
}

// removeStateTracking removes preconfirmed state tracking of the given message IDs, equivalent to
// marking the state as StateUnknown.
func (t *TxrV2) removeStateTracking(msgIDs ...string) {
	t.preconfirmedMu.Lock()
	defer t.preconfirmedMu.Unlock()

	for _, msgID := range msgIDs {
		delete(t.preconfirmedStates, msgID)
	}
}

// resendStaleTxns resends all the stale (pending) transactions in the tx pool.
func (t *TxrV2) resendStaleTxns(ctx context.Context) error {
	sCtx := sdk.UnwrapContext(ctx)
	chain := sCtx.Chain()

	content, err := chain.TxPoolContentFrom(ctx, t.signerAddr)
	if err != nil {
		t.logger.Error("failed to get tx pool content", "err", err)
		return err
	}

	for _, txn := range content["pending"] {
		bumpedTxn := sender.BumpGas(txn)
		if err = t.sender.SendTransaction(ctx, bumpedTxn); err != nil {
			t.logger.Error("failed to resend stale transaction", "err", err)
			return err
		}
	}
	return nil
}
