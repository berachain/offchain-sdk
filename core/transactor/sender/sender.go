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
		retryPolicy:         NewExponentialRetryPolicy(),
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
		// If sending the transaction fails, retry according to the retry policy, but only apply
		// this policy once.
		if shouldRetry {
			sCtx.Logger().Error("failed to send tx, retrying...", "hash", tx.Hash(), "err", err)
			go s.replaceTxWithPolicy(sCtx, tx, msgIDs, err)
		} else {
			// We were unable to send this tx, so let the factory know.
			s.factory.MarkTransactionNotSent(tx.Nonce())
		}
		return err
	}

	// If no error on sending, start tracking the inFlight transaction.
	s.tracker.Track(ctx, &tracker.InFlightTx{Transaction: tx, MsgIDs: msgIDs})
	return nil
}

// replaceTxWithPolicy retries sending tx according to the retry policy. Specifically handles two
// common errors on sending a transaction: NonceTooLow & ReplaceUnderpriced.
func (s *Sender) replaceTxWithPolicy(
	sCtx *sdk.Context, tx *coretypes.Transaction, msgIDs []string, err error,
) {
	for {
		retry, backoff := s.retryPolicy(sCtx, tx, err)
		if !retry {
			return
		}
		time.Sleep(backoff)

		// Bump the gas according to the replacement policy if a replacement is required.
		if errors.Is(err, txpool.ErrReplaceUnderpriced) ||
			(err != nil && strings.Contains(err.Error(), "replacement transaction underpriced")) {
			oldPrice := tx.GasPrice()
			tx = s.txReplacementPolicy(sCtx, tx)
			sCtx.Logger().Debug(
				"retrying with new gas and nonce",
				"old", oldPrice, "new", tx.GasPrice(), "nonce", tx.Nonce(),
			)
		}

		if errors.Is(err, core.ErrNonceTooLow) ||
			(err != nil && strings.Contains(err.Error(), "nonce too low")) {
			if tx, err = s.factory.BuildTransactionFromRequests(sCtx, &types.TxRequest{
				To:        tx.To(),
				Value:     tx.Value(),
				Data:      tx.Data(),
				Gas:       tx.Gas(),
				GasFeeCap: tx.GasFeeCap(),
				GasTipCap: tx.GasTipCap(),
				GasPrice:  tx.GasPrice(),
			}); err != nil {
				sCtx.Logger().Error("failed to build tx", "err", err)
				return
			}
		}

		// Sign the retry tx.
		tx, err = s.factory.SignTransaction(tx)
		if err != nil {
			sCtx.Logger().Error("failed to sign replacement transaction", "err", err)
			continue
		}

		// retry sending the transaction
		if err = s.SendTransactionAndTrack(sCtx, tx, msgIDs, false); err != nil {
			return
		}
	}
}
