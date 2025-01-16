package transactor

import (
	"context"
	"errors"
	"sort"
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

	requests     queuetypes.Queue[*types.Request]
	factory      *factory.Factory
	noncer       *tracker.Noncer
	sender       *sender.Sender
	senderMu     sync.Mutex
	dispatcher   *event.Dispatcher[*tracker.Response]
	tracker      *tracker.Tracker
	trackerIndex int
	chain        eth.Client

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
	factory := factory.New(noncer, batcher, signer, cfg.SignTxTimeout, cfg.MulticallRequireSuccess)
	dispatcher := event.NewDispatcher[*tracker.Response]()
	tracker := tracker.New(noncer, dispatcher, signer.Address(), cfg.TxWaitingTimeout)

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
	t.chain = chain
	t.logger = sCtx.Logger()

	// Register the transactor as a subscriber to the tracker.
	t.trackerIndex = t.SubscribeTxResults(ctx, t)

	// Setup and start all the transactor components.
	t.factory.SetClient(chain)
	t.sender.Setup(chain, t.logger)
	t.tracker.SetClient(chain)
	t.noncer.Start(ctx, chain)

	// If there are any pending txns at startup, they are likely to be "stuck". Resend them.
	if err := t.resendStaleTxns(ctx, chain); err != nil {
		return err
	}

	go t.mainLoop(ctx)

	return nil
}

// Execute implements job.Basic.
func (t *TxrV2) Execute(context.Context, any) (any, error) {
	acquired, inFlight := t.noncer.Stats()
	t.logger.Info(
		"🧠 system status",
		"waiting-tx", acquired, "in-flight-tx", inFlight, "pending-requests", t.requests.Len(),
	)
	return 1, nil
}

// IntervalTime implements job.Polling.
func (t *TxrV2) IntervalTime(context.Context) time.Duration {
	return t.cfg.StatusUpdateInterval
}

// Teardown implements job.HasTeardown.
func (t *TxrV2) Teardown() error {
	t.dispatcher.Unsubscribe(t.trackerIndex)
	return nil
}

// SubscribeTxResults ensures that tx results, once confirmed, are sent the given subscriber. It
// returns the global index of the subscription for the results.
func (t *TxrV2) SubscribeTxResults(ctx context.Context, subscriber tracker.Subscriber) int {
	ch := make(chan *tracker.Response)
	go tracker.NewSubscription(subscriber, t.logger).Start(ctx, ch)
	return t.dispatcher.Subscribe(ch)
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

// ForceTxRequest immediately (whenever the sender is free from any previous sends) builds and
// sends the tx request to the chain, after validating it.
// NOTE: this bypasses the queue and batching even if configured to do so.
func (t *TxrV2) ForceTxRequest(
	ctx context.Context, txReq *types.Request, async bool,
) (string, error) {
	if err := txReq.Validate(); err != nil {
		return "", err
	}

	if async {
		go t.fire(
			ctx,
			&tracker.Response{
				MsgIDs: []string{txReq.MsgID}, InitialTimes: []time.Time{txReq.Time()},
			},
			true, txReq.CallMsg,
		)
	} else {
		t.fire(
			ctx,
			&tracker.Response{
				MsgIDs: []string{txReq.MsgID}, InitialTimes: []time.Time{txReq.Time()},
			},
			true, txReq.CallMsg,
		)
	}
	return txReq.MsgID, nil
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

// resendStaleTxns resends all the stale (pending) transactions in the mempool with bumped gas.
// NOTE: blocks until resending all the pending txs either error and/or are sent to the chain.
func (t *TxrV2) resendStaleTxns(ctx context.Context, chain eth.Client) error {
	txPoolContent, err := chain.TxPoolContentFrom(ctx, t.signerAddr)
	if err != nil {
		t.logger.Error("failed to get tx pool content from", "err", err)
		return err
	}

	if pendingTxs := txPoolContent["pending"]; len(pendingTxs) > 0 {
		t.logger.Info("🔄 resending stale (pending in txpool) txs", "count", len(pendingTxs))

		// Create a sorted slice of nonces
		nonces := make([]uint64, 0, len(pendingTxs))
		for _, tx := range pendingTxs {
			nonces = append(nonces, tx.Nonce())
		}
		sort.Slice(nonces, func(i, j int) bool { return nonces[i] < nonces[j] })

		for _, nonce := range nonces {
			tx := pendingTxs[nonce]
			resp := &tracker.Response{Transaction: sender.BumpGas(tx, t.chain)}
			t.fire(ctx, resp, true, types.CallMsgFromTx(resp.Transaction))
			t.logger.Info("🔄 resending stale (pending in txpool) tx", "nonce", tx.Nonce())
		}
	}

	return nil
}
