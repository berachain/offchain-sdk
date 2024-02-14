package sender

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/berachain/offchain-sdk/core/transactor/tracker"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/txpool"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Tracker is an interface for tracking sent transactions.
type Tracker interface {
	Track(ctx context.Context, tx *tracker.InFlightTx)
}

// Factory is an interface for signing transactions.
type Factory interface {
	BuildTransactionFromRequests(
		ctx context.Context, txReqs ...*types.TxRequest,
	) (*coretypes.Transaction, error)
	SignTransaction(tx *coretypes.Transaction) (*coretypes.Transaction, error)
	MarkTransactionNotSent(nonce uint64)
}

// Sender struct holds the transaction replacement and retry policies.
type Sender struct {
	factory             Factory             // factory to sign new transactions
	tracker             Tracker             // tracker to track sent transactions
	txReplacementPolicy TxReplacementPolicy // policy to replace transactions
	retryPolicy         RetryPolicy         // policy to retry transactions
}

// New creates a new Sender with default replacement and exponential retry policies.
func New(factory Factory, tracker Tracker) *Sender {
	return &Sender{
		tracker:             tracker,
		factory:             factory,
		txReplacementPolicy: DefaultTxReplacementPolicy,
		retryPolicy:         &ExpoRetryPolicy{}, // TODO: choose from config.
	}
}

// SendTransaction sends a transaction using the Ethereum client. If the transaction fails,
// it retries based on the retry policy, only once (further retries will not retry again). If
// sending is successful, it uses the tracker to track the transaction.
func (s *Sender) SendTransactionAndTrack(
	ctx context.Context, tx *coretypes.Transaction, msgIDs []string, shouldRetry bool,
) error {
	sCtx := sdk.UnwrapContext(ctx)

	if err := sCtx.Chain().SendTransaction(ctx, tx); err != nil {
		// If sending the transaction fails, retry according to the retry policy.
		if shouldRetry {
			go s.retryTxWithPolicy(sCtx, tx, msgIDs, err)
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
	sCtx *sdk.Context, tx *coretypes.Transaction, msgIDs []string, err error,
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
		sCtx.Logger().Error("failed to send tx, retrying...", "hash", currTx, "err", err)

		// Bump the gas according to the replacement policy if a replacement is required.
		if errors.Is(err, txpool.ErrReplaceUnderpriced) ||
			(err != nil && strings.Contains(err.Error(), "replacement transaction underpriced")) {
			tx = s.txReplacementPolicy(sCtx, tx)
		}

		// Replace the nonce by asking the factory to rebuild this transaction.
		if errors.Is(err, core.ErrNonceTooLow) ||
			(err != nil && strings.Contains(err.Error(), "nonce too low")) {
			s.factory.MarkTransactionNotSent(currNonce)
			if tx, err = s.factory.BuildTransactionFromRequests(sCtx, &types.TxRequest{
				To:        tx.To(),
				Value:     tx.Value(),
				Data:      tx.Data(),
				Gas:       tx.Gas(),
				GasFeeCap: tx.GasFeeCap(),
				GasTipCap: tx.GasTipCap(),
				GasPrice:  tx.GasPrice(),
			}); err != nil {
				sCtx.Logger().Error("failed to build tx on replacing 'nonce too low'", "err", err)
				return
			}
		}

		// Update the retry policy if the transaction has been replaced and log.
		if newTx := tx.Hash(); newTx != currTx {
			sCtx.Logger().Debug(
				"retrying with diff gas and/or nonce",
				"old-gas", currGasPrice, "new-gas", tx.GasPrice(),
				"old-nonce", currNonce, "new-nonce", tx.Nonce(),
			)
			s.retryPolicy.updateTxReplacement(currTx, newTx)
		}

		// Sign the retry transaction.
		tx, err = s.factory.SignTransaction(tx)
		if err != nil {
			sCtx.Logger().Error("failed to sign replacement transaction", "err", err)
			continue
		}

		// Retry sending the transaction.
		err = s.SendTransactionAndTrack(sCtx, tx, msgIDs, false)
	}
}
