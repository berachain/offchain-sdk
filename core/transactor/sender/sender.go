package sender

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/core/transactor/tracker"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	"github.com/berachain/offchain-sdk/log"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Tracker is an interface for tracking sent transactions.
type Tracker interface {
	SetClient(chain eth.Client)
	Track(ctx context.Context, tx *tracker.InFlightTx)
}

// Factory is an interface for signing transactions.
type Factory interface {
	BuildTransactionFromRequests(
		ctx context.Context, forcedNonce uint64, txReqs ...*types.TxRequest,
	) (*coretypes.Transaction, error)
	GetNextNonce(oldNonce uint64) (uint64, bool)
}

// Sender struct holds the transaction replacement and retry policies.
type Sender struct {
	factory             Factory             // factory to sign new transactions
	tracker             Tracker             // tracker to track sent transactions
	txReplacementPolicy TxReplacementPolicy // policy to replace transactions
	retryPolicy         RetryPolicy         // policy to retry transactions
	chain               eth.Client
	logger              log.Logger
}

// New creates a new Sender with default replacement and exponential retry policies.
func New(factory Factory, tracker Tracker) *Sender {
	return &Sender{
		tracker:             tracker,
		factory:             factory,
		txReplacementPolicy: &DefaultTxReplacementPolicy{nf: factory},
		retryPolicy:         &ExpoRetryPolicy{}, // TODO: choose from config.
	}
}

func (s *Sender) Setup(chain eth.Client, logger log.Logger) {
	s.chain = chain
	s.logger = logger
	s.tracker.SetClient(chain)
}

// SendTransaction sends a transaction using the Ethereum client. If the transaction fails,
// it retries based on the retry policy, only once (further retries will not retry again). If
// sending is successful, it uses the tracker to track the transaction.
func (s *Sender) SendTransactionAndTrack(
	ctx context.Context, tx *coretypes.Transaction, msgIDs []string, shouldRetry bool,
) error {
	if err := s.chain.SendTransaction(ctx, tx); err != nil { // TODO: set timeout on context
		// If sending the transaction fails, retry according to the retry policy.
		if shouldRetry {
			go s.retryTxWithPolicy(ctx, tx, msgIDs, err)
		}
		return err
	}

	// If no error on sending, start tracking the inFlight transaction.
	s.tracker.Track(ctx, &tracker.InFlightTx{Transaction: tx, MsgIDs: msgIDs})
	return nil
}

// retryTxWithPolicy retries sending tx according to the retry policy. Specifically handles two
// common errors on sending a transaction (NonceTooLow, ReplaceUnderpriced) by replacing the tx
// appropriately.
func (s *Sender) retryTxWithPolicy(
	ctx context.Context, tx *coretypes.Transaction, msgIDs []string, err error,
) {
	for {
		// Check the policy to see if we should retry this transaction.
		retry, backoff := s.retryPolicy.get(tx, err)
		if !retry {
			return
		}
		time.Sleep(backoff) // Retry after recommended backoff.

		// Log relevant details about retrying the transaction.
		currTx, currGasPrice, currNonce := tx.Hash(), tx.GasPrice(), tx.Nonce()
		s.logger.Error("failed to send tx, retrying...", "hash", currTx, "err", err)

		// Get the replacement tx if necessary.
		tx = s.txReplacementPolicy.getNew(tx, err)

		// Update the retry policy if the transaction has been changed and log.
		if newTx := tx.Hash(); newTx != currTx {
			s.logger.Debug(
				"retrying with diff gas and/or nonce",
				"old-gas", currGasPrice, "new-gas", tx.GasPrice(),
				"old-nonce", currNonce, "new-nonce", tx.Nonce(),
			)
			s.retryPolicy.updateTxModified(currTx, newTx)
		}

		// Sign the retry transaction.
		if tx, err = s.factory.BuildTransactionFromRequests(ctx, tx.Nonce(), &types.TxRequest{
			To:        tx.To(),
			Gas:       tx.Gas(),
			GasPrice:  tx.GasPrice(),
			GasFeeCap: tx.GasFeeCap(),
			GasTipCap: tx.GasTipCap(),
			Value:     tx.Value(),
			Data:      tx.Data(),
		}); err != nil {
			s.logger.Error("failed to sign replacement transaction", "err", err)
			continue
		}

		// Retry sending the transaction.
		err = s.SendTransactionAndTrack(ctx, tx, msgIDs, false)
	}
}
