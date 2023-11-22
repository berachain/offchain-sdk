package sender

import (
	"context"
	"time"

	"github.com/berachain/offchain-sdk/core/transactor/tracker"
	sdk "github.com/berachain/offchain-sdk/types"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Factory interface for signing transactions.
type Factory interface {
	SignTransaction(*coretypes.Transaction) (*coretypes.Transaction, error)
}

// Sender struct holds the transaction replacement and retry policies.
type Sender struct {
	factory             Factory
	txReplacementPolicy TxReplacementPolicy // policy to replace transactions
	retryPolicy         RetryPolicy         // policy to retry transactions
}

// New creates a new Sender with default replacement and retry policies.
func New(factory Factory) *Sender {
	return &Sender{
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

	// Send the replacement transaction.
	// TODO: needs to be resigned by factory.
	if err := ethClient.SendTransaction(ctx, tx); err != nil { // if sending the transaction fails
		sCtx.Logger().Error(
			"failed to resend replacement transaction", "hash", tx.Hash(), "err", err) // log the error
		if retry, backoff := s.retryPolicy(ctx, tx, err); retry {
			time.Sleep(backoff)                               // wait for the backoff time
			if err = s.SendTransaction(ctx, tx); err != nil { // retry sending the transaction
				return err // if it fails again, return the error
			}
		}
		return err // if the retry policy does not allow for a retry, return the error
	}
	return nil // if the transaction was sent successfully, return nil
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
func (s *Sender) OnError(ctx context.Context, tx *tracker.InFlightTx, _ error) {
	replacementTx, err := s.factory.SignTransaction(s.txReplacementPolicy(ctx, tx.Transaction))
	if err != nil {
		sdk.UnwrapContext(ctx).Logger().Error(
			"failed to sign replacement transaction", "err", err)
		return
	}
	_ = s.SendTransaction(ctx, replacementTx)
}
