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
		noncer:              noncer,                      // noncer to acquire nonces
		factory:             factory,                     // factory to sign transactions
		txReplacementPolicy: DefaultTxReplacementPolicy,  // default tx replacement policy
		retryPolicy:         NewExponentialRetryPolicy(), // exponential backoff retry policy
	}
}

// SendTransaction sends a transaction using the Ethereum client. If the transaction fails,
// it retries based on the retry policy.
func (s *Sender) SendTransaction(ctx context.Context, tx *coretypes.Transaction) error {
	sCtx := sdk.UnwrapContext(ctx) // unwrap the context to get the SDK context
	ethClient := sCtx.Chain()      // get the Ethereum client from the SDK context

	if err := ethClient.SendTransaction(ctx, tx); err != nil { // if sending the transaction fails
		sCtx.Logger().Error(
			"failed to send tx", "hash", tx.Hash(), "err", err, // log the error
		)
		go s.retryTxWithPolicy(sCtx, tx, err) // retry according to the retry policy
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
	return s.retryTx(sdk.UnwrapContext(ctx), tx.Transaction)
}

// OnError is called when an error occurs while sending a transaction. In this case, the
// transaction is replaced with a new transaction with a higher gas price as defined by
// the txReplacementPolicy.
func (s *Sender) OnError(ctx context.Context, tx *tracker.InFlightTx, err error) {
	sCtx := sdk.UnwrapContext(ctx)

	// Assign the new transaction to the in-flight transaction.
	tx.Transaction = s.handleNonceTooLow(sCtx, tx.Transaction, err)
	tx.Receipt = nil

	// The original tx was never sent so we remove from the in-flight list.
	s.noncer.RemoveInFlight(tx)

	_ = s.retryTx(sCtx, tx.Transaction)
}

// retryTxWithPolicy retries the tx according to the retry policy. If the nonce is too low, builds
// a new tx with the latest nonce from the factory & noncer.
func (s *Sender) retryTxWithPolicy(sCtx *sdk.Context, tx *coretypes.Transaction, err error) {
	tx = s.handleNonceTooLow(sCtx, tx, err)

	for {
		retry, backoff := s.retryPolicy(sCtx, tx, err)
		if !retry {
			return
		}

		time.Sleep(backoff) // wait for the backoff time
		err = s.retryTx(sCtx, tx)
	}
}

// retryTx manages the logic for replacing a tx according to the replacement policy and resending.
func (s *Sender) retryTx(sCtx *sdk.Context, tx *coretypes.Transaction) error {
	replacementTx := s.txReplacementPolicy(sCtx, tx)
	sCtx.Logger().Debug(
		"retrying with new gas and nonce",
		"old", tx.GasPrice(), "new", replacementTx.GasPrice(), "nonce", tx.Nonce(),
	)

	// sign the tx with the new gas price
	signedTx, err := s.factory.SignTransaction(replacementTx)
	if err != nil {
		sCtx.Logger().Error("failed to sign replacement transaction", "err", err)
		return err
	}

	// retry sending the transaction
	return s.SendTransaction(sCtx, signedTx)
}

// handleNonceTooLow will replace a transaction with a new one if the nonce is too low.
func (s *Sender) handleNonceTooLow(
	sCtx *sdk.Context, tx *coretypes.Transaction, err error,
) *coretypes.Transaction {
	if !(errors.Is(err, core.ErrNonceTooLow) || strings.Contains(err.Error(), "nonce too low")) {
		return tx
	}

	ethTx, buildErr := s.factory.BuildTransaction(sCtx, &types.TxRequest{
		To:        tx.To(),
		Gas:       tx.Gas(),
		GasPrice:  tx.GasPrice(),
		GasFeeCap: tx.GasFeeCap(),
		GasTipCap: tx.GasTipCap(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	})
	if buildErr != nil {
		sCtx.Logger().Error("failed to build replacement transaction", "err", err)
		return tx
	}

	return ethTx
}
