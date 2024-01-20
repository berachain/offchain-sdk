package factory

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/berachain/offchain-sdk/contracts/bindings"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

const (
	tryAggregate      = `tryAggregate`
	executionReverted = `execution reverted: `
)

// forge create Multicall3 --rpc-url=http://devnet.beraswillmakeit.com:8545
// --private-key=0xfffdbb37105441e14b0ee6330d855d8504ff39e705c3afa8f859ac9865f99306.
type Multicall3Batcher struct {
	contractAddress common.Address
	packer          *types.Packer
}

// NewMulticall3Batcher creates a new Multicall3Batcher instance.
func NewMulticall3Batcher(address common.Address) *Multicall3Batcher {
	return &Multicall3Batcher{
		contractAddress: address,
		packer:          &types.Packer{MetaData: bindings.Multicall3MetaData},
	}
}

// BatchTxRequests creates a batched transaction request for the given transaction requests.
func (mc *Multicall3Batcher) BatchTxRequests(
	_ context.Context, txReqs ...*types.TxRequest,
) *types.TxRequest {
	var (
		calls       = make([]bindings.Multicall3Call, len(txReqs))
		totalValue  = big.NewInt(0)
		gasLimit    = uint64(0)
		gasTipCap   *big.Int
		gasFeeCap   *big.Int
		gasPriceSet = false
	)

	for i, txReq := range txReqs {
		// use the summed value for the batched transaction.
		if txReq.Value != nil {
			totalValue = totalValue.Add(totalValue, txReq.Value)
		}

		// use the summed gas limit for the batched transaction.
		gasLimit += txReq.Gas

		// set the gas prices to the first non-nil gas prices in the batch.
		if !gasPriceSet {
			gasTipCap = txReq.GasTipCap
			gasFeeCap = txReq.GasFeeCap
			gasPriceSet = true
		}

		call := bindings.Multicall3Call{
			Target:   *txReq.To,
			CallData: txReq.Data,
		}
		calls[i] = call
	}

	txRequest, _ := mc.packer.CreateTxRequest(
		mc.contractAddress, totalValue, gasTipCap, gasFeeCap, gasLimit, tryAggregate, false, calls,
	)
	return txRequest
}

// BatchCallRequests uses the Multicall3 contract to create a batched call request for the given
// tx requests and return the batched call response data for each call.
func (mc *Multicall3Batcher) BatchCallRequests(
	ctx context.Context,
	from common.Address,
	txReqs ...*types.TxRequest,
) ([]bindings.Multicall3Result, error) {
	sCtx := sdk.UnwrapContext(ctx)

	// get the batched tx (call) requests
	batchedCall := mc.BatchTxRequests(ctx, txReqs...)
	batchedCall.From = from

	// call the multicall3 contract with the batched call request
	ret, err := sCtx.Chain().CallContract(ctx, ethereum.CallMsg(*batchedCall), nil)
	if err != nil {
		if _, reason, ok := strings.Cut(err.Error(), executionReverted); ok {
			sCtx.Logger().Warn("execution reverted for multicall3", "reason", reason)
		} else {
			sCtx.Logger().Error("failed to call contract", "err", err)
		}
		return nil, err
	}

	// unpack the return data into call responses
	callResult, err := mc.packer.GetCallResponse(tryAggregate, ret)
	if err != nil {
		sCtx.Logger().Error("failed to unpack call response", "err", err)
		return nil, err
	}
	if len(callResult) != 1 {
		err = fmt.Errorf("expected 1 list of Multicall3Results, got %d", len(callResult))
		sCtx.Logger().Error("failed to unpack call response", "err", err)
		return nil, err
	}
	callResults, ok := callResult[0].([]struct {
		Success    bool    "json:\"success\""
		ReturnData []uint8 "json:\"returnData\""
	})
	if !ok {
		err = errors.New("expected return type as list of Multicall3Results")
		sCtx.Logger().Error("failed to unpack call response", "err", err)
		return nil, err
	}

	// convert the call responses into Multicall3Results
	multicall3Results := make([]bindings.Multicall3Result, len(callResults))
	for i, callResult := range callResults {
		multicall3Results[i] = bindings.Multicall3Result{
			Success:    callResult.Success,
			ReturnData: callResult.ReturnData,
		}
	}
	return multicall3Results, nil
}
