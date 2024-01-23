package sender

import (
	"context"
	"errors"
	"time"

	"github.com/berachain/offchain-sdk/core/transactor/tracker"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/ethereum/go-ethereum/core"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

type Noncer interface {
	RemoveInFlight(tx *tracker.InFlightTx)
}

// Factory interface for signing transactions.
type Factory interface {
	BuildTransaction(context.Context, *types.TxRequest) (*coretypes.Transaction, error)
	SignTransaction(*coretypes.Transaction) (*coretypes.Transaction, error)
}

// Sender struct holds the transaction replacement and retry policies.
type Sender struct {
	noncer              Noncer // noncer to acquire nonces
	factory             Factory
	txReplacementPolicy TxReplacementPolicy // policy to replace transactions
	retryPolicy         RetryPolicy         // policy to retry transactions
}

// New creates a new Sender with default replacement and retry policies.
func New(factory Factory, noncer Noncer) *Sender {
	return &Sender{
		noncer:              noncer,                     // noncer to acquire nonces
		factory:             factory,                    // factory to sign transactions
		txReplacementPolicy: DefaultTxReplacementPolicy, // default transaction replacement policy
		retryPolicy:         DefaultRetryPolicy,         // default retry policy
	}
}

// SendTransaction sends a transaction using the Ethereum client. If the transaction fails,
// it retries based on the retry policy.
func (s *Sender) SendTransaction(ctx context.Context, tx *coretypes.Transaction) error {
	sCtx := sdk.UnwrapContext(ctx) // unwrap the context to get the SDK context
	ethClient := sCtx.Chain()      // get the Ethereum client from the SDK context

	// TODO: needs to be resigned by factory.
	// TODO: this returns the revert error message, handle it.
	if err := ethClient.SendTransaction(ctx, tx); err != nil { // if sending the transaction fails
		sCtx.Logger().Error(
			"failed to send tx transaction", "hash", tx.Hash(), "err", err, // log the error
		)

		// Send the replacement transaction.
		price := tx.GasPrice()
		tx = s.txReplacementPolicy(ctx, tx)
		sCtx.Logger().Info(
			"retrying with new gas limit", "old", price, "new", tx.GasPrice(), "nonce", tx.Nonce(),
		)
		if retry, backoff := s.retryPolicy(ctx, tx, err); retry {
			time.Sleep(backoff)                               // wait for the backoff time
			if err = s.SendTransaction(ctx, tx); err != nil { // retry sending the transaction
				return err // if it fails again, return the error
			}
		}

		// if the retry policy does not allow for a retry, return the error
		return err
	}

	// if the transaction was sent successfully, return nil
	return nil
}

// On Success for the sender is a no-op since there is nothing else to do if the transaction
// is successful.
func (s *Sender) OnSuccess(*tracker.InFlightTx, *coretypes.Receipt) error {
	return nil
}

// OnRevert is called when a transaction is reverted, for the sender this is also currently a
// no-op.
func (s *Sender) OnRevert(*tracker.InFlightTx, *coretypes.Receipt) error {
	return nil
}

// OnStale is called when a transaction is marked as stale by the tracker. In this case, the
// transaction is replaced with a new transaction with a higher gas price as defined by the
// txReplacementPolicy.
func (s *Sender) OnStale(ctx context.Context, tx *tracker.InFlightTx) error {
	replacementTx, err := s.factory.SignTransaction(s.txReplacementPolicy(ctx, tx.Transaction))
	if err != nil {
		sdk.UnwrapContext(ctx).Logger().Error(
			"failed to sign replacement transaction", "err", err)
		return err
	}
	return s.SendTransaction(ctx, replacementTx)
}

// OnError is called when an error occurs while sending a transaction. In this case, the
// transaction is replaced with a new transaction with a higher gas price as defined by
// the txReplacementPolicy.
// TODO: make this more robust probably.
func (s *Sender) OnError(ctx context.Context, tx *tracker.InFlightTx, err error) {
	if errors.Is(err, core.ErrNonceTooLow) {
		ethTx, buildErr := s.factory.BuildTransaction(ctx, &types.TxRequest{
			To:    tx.To(),
			Value: tx.Value(),
			Data:  tx.Data(),
		})
		if buildErr != nil {
			sdk.UnwrapContext(ctx).Logger().Error(
				"failed to build replacement transaction", "err", err)
			return
		}
		// The original tx was never sent so we remove from the in-flight list.
		s.noncer.RemoveInFlight(tx)

		// Assign the new transaction to the in-flight transaction.
		tx.Transaction = ethTx
		tx.Receipt = nil
	}

	replacementTx, err := s.factory.SignTransaction(s.txReplacementPolicy(ctx, tx.Transaction))
	if err != nil {
		sdk.UnwrapContext(ctx).Logger().Error(
			"failed to sign replacement transaction", "err", err)
		return
	}
	if err = s.SendTransaction(ctx, replacementTx); err != nil {
		sdk.UnwrapContext(ctx).Logger().Error(
			"failed to send replacement transaction", "err", err)
		return
	}
}
