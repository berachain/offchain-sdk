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

// awsMaxBatchSize is the max batch size for AWS.
const awsMaxBatchSize = 10

// TxrV2 is the main transactor object.
type TxrV2 struct {
	cfg        Config
	requests   queuetypes.Queue[*types.TxRequest]
	sender     *sender.Sender
	factory    *factory.Factory
	noncer     *tracker.Noncer
	tracker    *tracker.Tracker
	dispatcher *event.Dispatcher[*tracker.InFlightTx]
	logger     log.Logger
	mu         sync.Mutex
}

// NewTransactor creates a new transactor with the given config, request queue
// and signer.
func NewTransactor(
	cfg Config, queue queuetypes.Queue[*types.TxRequest], signer kmstypes.TxSigner,
) *TxrV2 {
	noncer := tracker.NewNoncer(signer.Address())
	factory := factory.New(
		noncer, signer,
		factory.NewMulticall3Batcher(common.HexToAddress(cfg.Multicall3Address)),
	)

	dispatcher := event.NewDispatcher[*tracker.InFlightTx]()

	txr := &TxrV2{
		dispatcher: dispatcher,
		cfg:        cfg,
		factory:    factory,
		sender:     sender.New(factory, noncer),
		noncer:     noncer,
		tracker:    tracker.New(noncer, dispatcher, cfg.TxReceiptTimeout),
		requests:   queue,
		mu:         sync.Mutex{},
	}

	// Register the tracker as a subscriber to the tracker.
	ch := make(chan *tracker.InFlightTx, 1024) //nolint:gomnd // its okay.
	go func() {
		// TODO: handle error
		_ = tracker.NewSubscription(txr, txr.logger).Start(context.Background(), ch)
	}()
	dispatcher.Subscribe(ch)

	// Register the sender as a subscriber to the tracker.
	ch2 := make(chan *tracker.InFlightTx, 1024) //nolint:gomnd // its okay.
	go func() {
		// TODO: handle error
		_ = tracker.NewSubscription(txr.sender, txr.logger).Start(context.Background(), ch2)
	}()
	dispatcher.Subscribe(ch2)

	return txr
}

// RegistryKey implements job.Basic.
func (t *TxrV2) RegistryKey() string {
	return "transactor"
}

// SubscribeTxResults sends the tx results (inflight) to the given channel.
func (t *TxrV2) SubscribeTxResults(
	ctx context.Context, subscriber tracker.Subscriber, ch chan *tracker.InFlightTx,
) {
	go func() {
		// TODO: handle error
		_ = tracker.NewSubscription(subscriber, t.logger).Start(ctx, ch)
	}()
	t.dispatcher.Subscribe(ch)
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
	return 5 * time.Second //nolint:gomnd // its okay.
}

// Setup implements job.HasSetup.
// TODO: deprecate off being a job.
func (t *TxrV2) Setup(ctx context.Context) error {
	// todo: need lock on nonce to support more than one
	t.logger = sdk.UnwrapContext(ctx).Logger()
	t.noncer.SetClient(sdk.UnwrapContext(ctx).Chain())
	t.Start(sdk.UnwrapContext(ctx))
	return nil
}

// SendTxRequest adds the given tx request to the tx queue.
func (t *TxrV2) SendTxRequest(txReq *types.TxRequest) (string, error) {
	return t.requests.Push(txReq)
}

// Start starts the transactor.
func (t *TxrV2) Start(ctx context.Context) {
	go t.mainLoop(ctx)
	go t.noncer.RefreshLoop(ctx)
}

// mainLoop is the main transaction sending / batching loop.
func (t *TxrV2) mainLoop(ctx context.Context) {
	if err := t.noncer.InitializeExistingTxs(ctx); err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Attempt the retrieve a batch from the queue.
			msgIDs, batch := t.retrieveBatch(ctx)

			// We didn't get any transactions, so we wait for more.
			if len(batch) == 0 {
				t.logger.Info("no tx requests to process....")
				time.Sleep(t.cfg.EmtpyQueueDelay)
				continue
			}

			// We got a batch, so we send it and track it.
			// We must first wait for the previous sending to finish.
			t.mu.Lock()
			go func() {
				defer t.mu.Unlock()
				if err := t.sendAndTrack(ctx, msgIDs, batch...); err != nil {
					t.logger.Error("failed to process batch", "err", err)
				}
			}()
		}
	}
}

// retrieveBatch retrieves a batch of transaction requests from the queue.
// It waits until it hits the max batch size or the timeout.
func (t *TxrV2) retrieveBatch(_ context.Context) ([]string, []*types.TxRequest) {
	var batch []*types.TxRequest
	var retMsgIDs []string
	startTime := time.Now()

	// Retrieve the smaller of the aws max batch size or the delta between the max total batch size.
	for len(batch) < t.cfg.TxBatchSize && time.Since(startTime) < t.cfg.TxBatchTimeout {
		msgIDs, txReq, err := t.requests.ReceiveMany(
			int32(min(awsMaxBatchSize, t.cfg.TxBatchSize-len(batch))),
		)
		if err != nil {
			t.logger.Error("failed to receive tx request", "err", err)
			continue
		}
		batch = append(batch, txReq...)
		retMsgIDs = append(retMsgIDs, msgIDs...)
	}
	return retMsgIDs, batch
}

// sendAndTrack processes a batch of transaction requests.
// It builds a transaction from the batch and sends it.
// It also tracks the transaction for future reference.
func (t *TxrV2) sendAndTrack(
	ctx context.Context, msgIDs []string, batch ...*types.TxRequest,
) error {
	tx, err := t.factory.BuildTransactionFromRequests(ctx, batch...)
	if err != nil {
		return err
	}

	// Send the transaction to the chain.
	if err = t.sender.SendTransaction(ctx, tx); err != nil {
		return err
	}

	// t.logger.Debug("📡 sent transaction", "tx-hash", tx.Hash().Hex(), "tx-reqs", len(batch))

	// Spin off a goroutine to track the transaction.
	t.tracker.Track(
		ctx,
		&tracker.InFlightTx{
			Transaction: tx,
			MsgIDs:      msgIDs,
		},
		true,
	)
	return nil
}
