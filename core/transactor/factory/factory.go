package factory

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/core/transactor/sender"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	kmstypes "github.com/berachain/offchain-sdk/types/kms/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

const signTxTimeout = 2 * time.Second // TODO: read from config.

// Factory is a transaction factory that builds 1559 transactions with the configured signer.
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

// BuildTransactionFromRequests builds a transaction from a list of requests. A non-zero nonce
// should only be provided if this is a retry with a specific nonce necessary.
func (f *Factory) BuildTransactionFromRequests(
	ctx context.Context, txReqs ...*types.TxRequest,
) (*types.BatchRequest, error) {
	switch len(txReqs) {
	case 0:
		return nil, errors.New("no transaction requests provided")
	case 1:
		// if len(txReqs) == 1 then build a single transaction.
		tx, err := f.buildTransaction(ctx, txReqs[0].CallMsg, 0)
		if err != nil {
			return nil, err
		}

		return &types.BatchRequest{
			Transaction: tx,
			MsgIDs:      []string{txReqs[0].MsgID},
			TimesFired:  []time.Time{txReqs[0].Time()},
		}, nil
	default:
		// len(txReqs) > 1 then build a multicall transaction.
		ar := f.mc3Batcher.BatchTxRequests(txReqs...)

		// Build the transaction to include the calldata.
		// ar.To should be the Multicall3 contract address
		// ar.Data should be the calldata with the batched transactions.
		// ar.Value is the sum of the values of the batched transactions.
		tx, err := f.buildTransaction(ctx, ar.CallMsg, 0)
		if err != nil {
			return nil, err
		}

		batch := &types.BatchRequest{
			Transaction: tx,
			MsgIDs:      make([]string, len(txReqs)),
			TimesFired:  make([]time.Time, len(txReqs)),
		}
		for i, txReq := range txReqs {
			batch.MsgIDs[i] = txReq.MsgID
			batch.TimesFired[i] = txReq.Time()
		}
		return batch, nil
	}
}

// RebuildBatch rebuilds an already batched transaction with the forced nonce.
func (f *Factory) RebuildBatch(
	ctx context.Context, batch *types.BatchRequest, forcedNonce uint64,
) (*types.BatchRequest, error) {
	var err error
	batch.Transaction, err = f.buildTransaction(ctx, types.NewCallMsgFromTx(batch), forcedNonce)
	return batch, err
}

// buildTransaction builds a transaction with the configured signer. If nonce of 0 is provided,
// a fresh nonce is acquired from the noncer.
func (f *Factory) buildTransaction(
	ctx context.Context, callMsg *ethereum.CallMsg, nonce uint64,
) (*coretypes.Transaction, error) {
	var err error

	// get the chain ID
	if f.chainID == nil {
		f.chainID, err = f.ethClient.ChainID(ctx)
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
		To:      callMsg.To,
		Value:   callMsg.Value,
		Data:    callMsg.Data,
		Nonce:   nonce,
	}

	// set gas tip cap from eth client if not already provided
	if callMsg.GasTipCap != nil {
		txData.GasTipCap = callMsg.GasTipCap
	} else {
		txData.GasTipCap, err = f.ethClient.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, err
		}
	}

	// set gas fee cap as (gasTipCap + 2 * basefee) if not already provided
	if callMsg.GasFeeCap != nil {
		txData.GasFeeCap = callMsg.GasFeeCap
	} else {
		head, err := f.ethClient.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, err
		}

		// use base fee wiggle multiplier of 2
		txData.GasFeeCap = new(big.Int).Add(
			txData.GasTipCap, new(big.Int).Mul(head.BaseFee, common.Big2),
		)
	}

	// set gas limit from eth client if not already provided
	if callMsg.Gas > 0 {
		txData.Gas = callMsg.Gas
	} else {
		callMsg.From = f.signer.Address() // set the from address for estimate gas
		if txData.Gas, err = f.ethClient.EstimateGas(ctx, *callMsg); err != nil {
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
	ctxWithTimeout, cancel := context.WithTimeout(ctx, signTxTimeout)
	signer, err := f.signer.SignerFunc(ctxWithTimeout, tx.ChainId())
	cancel()
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
