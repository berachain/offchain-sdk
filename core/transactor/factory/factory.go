package factory

import (
	"context"
	"errors"
	"math/big"

	"github.com/berachain/offchain-sdk/core/transactor/sender"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	sdk "github.com/berachain/offchain-sdk/types"
	kmstypes "github.com/berachain/offchain-sdk/types/kms/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Noncer is an interface for acquiring nonces.
type Noncer interface {
	Acquire(context.Context) (uint64, bool)
	RemoveAcquired(uint64)
}

// Factory is a transaction factory that builds transactions with the configured signer.
type Factory struct {
	noncer        Noncer
	signer        kmstypes.TxSigner
	signerAddress common.Address
	mc3Batcher    *Multicall3Batcher

	// caches
	chainID *big.Int
}

// New creates a new factory instance.
func New(noncer Noncer, signer kmstypes.TxSigner, mc3Batcher *Multicall3Batcher) *Factory {
	return &Factory{
		noncer:        noncer,
		signer:        signer,
		signerAddress: signer.Address(),
		mc3Batcher:    mc3Batcher,
	}
}

// BuildTransactionFromRequests builds a transaction from a list of requests.
func (f *Factory) BuildTransactionFromRequests(
	ctx context.Context, txReqs ...*types.TxRequest,
) (*coretypes.Transaction, error) {
	switch len(txReqs) {
	case 0:
		return nil, errors.New("no transaction requests provided")
	case 1:
		// if len(txReqs) == 1 then build a single transaction.
		return f.buildTransaction(ctx, txReqs[0])
	default:
		// len(txReqs) > 1 then build a multicall transaction.
		ar := f.mc3Batcher.BatchTxRequests(ctx, txReqs...)

		// Build the transaction to include the calldata.
		// ar.To should be the Multicall3 contract address
		// ar.Data should be the calldata with the batched transactions.
		// ar.Value is the sum of the values of the batched transactions.
		return f.buildTransaction(ctx, ar)
	}
}

// buildTransaction builds a transaction with the configured signer.
func (f *Factory) buildTransaction(
	ctx context.Context, txReq *types.TxRequest,
) (*coretypes.Transaction, error) {
	var (
		ethClient = sdk.UnwrapContext(ctx).Chain()
		err       error
	)

	// get the chain ID
	if f.chainID == nil {
		f.chainID, err = ethClient.ChainID(ctx)
		if err != nil {
			return nil, err
		}
	}

	// get the nonce from the noncer
	nonce, isReplacing := f.noncer.Acquire(ctx)
	defer func() {
		if err != nil {
			f.MarkTransactionNotSent(nonce)
		}
	}()

	// start building the 1559 transaction
	txData := &coretypes.DynamicFeeTx{
		ChainID: f.chainID,
		To:      txReq.To,
		Value:   txReq.Value,
		Data:    txReq.Data,
		Nonce:   nonce,
	}

	// set gas fee cap from eth client if not already provided
	if txReq.GasFeeCap != nil {
		txData.GasFeeCap = txReq.GasFeeCap
	} else {
		txData.GasFeeCap, err = ethClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
	}

	// set gas tip cap from eth client if not already provided
	if txReq.GasTipCap != nil {
		txData.GasTipCap = txReq.GasTipCap
	} else {
		txData.GasTipCap, err = ethClient.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, err
		}
	}

	// set gas limit from eth client if not already provided
	if txReq.Gas > 0 {
		txData.Gas = txReq.Gas
	} else {
		if txData.Gas, err = ethClient.EstimateGas(ctx, ethereum.CallMsg(*txReq)); err != nil {
			return nil, err
		}
	}

	// bump gas (if necessary) and sign the transaction.
	tx := coretypes.NewTx(txData)
	if isReplacing {
		tx = sender.DefaultTxReplacementPolicy(ctx, tx)
	}
	tx, err = f.SignTransaction(tx)
	return tx, err
}

// signTransaction signs a transaction with the configured signer.
func (f *Factory) SignTransaction(tx *coretypes.Transaction) (*coretypes.Transaction, error) {
	signer, err := f.signer.SignerFunc(context.Background(), tx.ChainId())
	if err != nil {
		return nil, err
	}
	return signer(f.signerAddress, tx)
}

// MarkTransactionNotSent lets the noncer know that the acquired nonce could not be sent.
func (f *Factory) MarkTransactionNotSent(nonce uint64) {
	f.noncer.RemoveAcquired(nonce)
}
