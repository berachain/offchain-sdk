package factory

import (
	"context"
	"errors"
	"math/big"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/core/transactor/sender"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	kmstypes "github.com/berachain/offchain-sdk/types/kms/types"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Factory is a transaction factory that builds transactions with the configured signer.
type Factory struct {
	noncer        Noncer
	signer        kmstypes.TxSigner
	signerAddress common.Address
	mc3Batcher    *Multicall3Batcher

	// caches
	ethClient eth.Client
	chainID   *big.Int
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

func (f *Factory) SetClient(ethClient eth.Client) {
	f.ethClient = ethClient
}

// BuildTransactionFromRequests builds a transaction from a list of requests.
func (f *Factory) BuildTransactionFromRequests(
	ctx context.Context, forcedNonce uint64, txReqs ...*types.TxRequest,
) (*coretypes.Transaction, error) {
	switch len(txReqs) {
	case 0:
		return nil, errors.New("no transaction requests provided")
	case 1:
		// if len(txReqs) == 1 then build a single transaction.
		return f.buildTransaction(ctx, forcedNonce, txReqs[0])
	default:
		// len(txReqs) > 1 then build a multicall transaction.
		ar := f.mc3Batcher.BatchTxRequests(txReqs...)

		// Build the transaction to include the calldata.
		// ar.To should be the Multicall3 contract address
		// ar.Data should be the calldata with the batched transactions.
		// ar.Value is the sum of the values of the batched transactions.
		return f.buildTransaction(ctx, forcedNonce, ar)
	}
}

// buildTransaction builds a transaction with the configured signer.
func (f *Factory) buildTransaction(
	ctx context.Context, nonce uint64, txReq *types.TxRequest,
) (*coretypes.Transaction, error) {
	var err error

	// get the chain ID
	if f.chainID == nil {
		f.chainID, err = f.ethClient.ChainID(ctx) // TODO: set timeout on context
		if err != nil {
			return nil, err
		}
	}

	// get the nonce from the noncer if not provided
	var isReplacing bool
	if nonce == 0 {
		nonce, isReplacing = f.noncer.Acquire()
		defer func() {
			if err != nil {
				f.noncer.RemoveAcquired(nonce)
			}
		}()
	}

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
		txData.GasFeeCap, err = f.ethClient.SuggestGasPrice(ctx) // TODO: set timeout on context
		if err != nil {
			return nil, err
		}
	}

	// set gas tip cap from eth client if not already provided
	if txReq.GasTipCap != nil {
		txData.GasTipCap = txReq.GasTipCap
	} else {
		txData.GasTipCap, err = f.ethClient.SuggestGasTipCap(ctx) // TODO: set timeout on context
		if err != nil {
			return nil, err
		}
	}

	// set gas limit from eth client if not already provided
	if txReq.Gas > 0 {
		txData.Gas = txReq.Gas
	} else {
		// TODO: set timeout on context
		if txData.Gas, err = f.ethClient.EstimateGas(ctx, *txReq.CallMsg); err != nil {
			return nil, err
		}
	}

	// bump gas (if necessary) and sign the transaction.
	tx := coretypes.NewTx(txData)
	if isReplacing {
		tx = sender.BumpGas(tx)
	}
	tx, err = f.SignTransaction(ctx, tx)
	return tx, err
}

// signTransaction signs a transaction with the configured signer.
func (f *Factory) SignTransaction(
	ctx context.Context, tx *coretypes.Transaction,
) (*coretypes.Transaction, error) {
	signer, err := f.signer.SignerFunc(ctx, tx.ChainId()) // TODO: set timeout on context
	if err != nil {
		return nil, err
	}
	return signer(f.signerAddress, tx)
}

// GetNextNonce lets the noncer know that the old nonce could not be sent and acquires a new one.
func (f *Factory) GetNextNonce(oldNonce uint64) (uint64, bool) {
	f.noncer.RemoveAcquired(oldNonce)
	return f.noncer.Acquire()
}
