package batcher

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/berachain/offchain-sdk/contracts/bindings"
	"github.com/berachain/offchain-sdk/core/transactor/factory"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

const (
	tryAggregate      = `tryAggregate`
	executionReverted = `execution reverted: `
)

var _ factory.Batcher = (*Multicall3)(nil)

// Corresponds to the Multicall3 contract (https://www.multicall3.com), also dumped into
// contracts/src/Multicall3.sol.
type Multicall3 struct {
	contractAddress common.Address
	packer          *types.Packer
}

// NewMulticall3 creates a new Multicall3 instance.
func NewMulticall3(address common.Address) *Multicall3 {
	return &Multicall3{
		contractAddress: address,
		packer:          &types.Packer{MetaData: bindings.Multicall3MetaData},
	}
}

// BatchRequests creates a batched transaction request for the given call requests.
func (mc *Multicall3) BatchRequests(
	requireSuccess bool, callReqs ...*ethereum.CallMsg,
) *types.Request {
	var (
		calls       = make([]bindings.Multicall3Call, len(callReqs))
		totalValue  = big.NewInt(0)
		gasLimit    = uint64(0)
		gasTipCap   *big.Int
		gasFeeCap   *big.Int
		gasPriceSet = false
	)

	for i, callReq := range callReqs {
		// use the summed value for the batched transaction.
		if callReq.Value != nil {
			totalValue = totalValue.Add(totalValue, callReq.Value)
		}

		// use the summed gas limit for the batched transaction.
		gasLimit += callReq.Gas

		// set the gas prices to the first non-nil gas prices in the batch.
		if !gasPriceSet {
			gasTipCap = callReq.GasTipCap
			gasFeeCap = callReq.GasFeeCap
			gasPriceSet = true
		}

		calls[i] = bindings.Multicall3Call{
			Target:   *callReq.To,
			CallData: callReq.Data,
		}
	}

	txRequest, _ := mc.packer.CreateRequest(
		"", mc.contractAddress, totalValue, gasTipCap, gasFeeCap, gasLimit,
		tryAggregate, requireSuccess, calls,
	)
	return txRequest
}

// BatchCallRequests uses the Multicall3 contract to create a batched call request for the given
// call messages and return the batched call result data for each call, as a `[]Multicall3Result`.
func (mc *Multicall3) BatchCallRequests(
	ctx context.Context, from common.Address, requireSuccess bool, callReqs ...*ethereum.CallMsg,
) (any, error) {
	sCtx := sdk.UnwrapContext(ctx)

	// get the batched tx (call) requests
	batchedCall := mc.BatchRequests(requireSuccess, callReqs...)
	batchedCall.From = from

	// call the multicall3 contract with the batched call request
	ret, err := sCtx.Chain().CallContract(ctx, *batchedCall.CallMsg, nil)
	if err != nil {
		if _, reason, ok := strings.Cut(err.Error(), executionReverted); ok {
			sCtx.Logger().Warn("execution reverted for multicall3", "reason", reason)
		} else {
			sCtx.Logger().Error("failed to call contract", "err", err)
		}
		return nil, err
	}

	// unpack the return data into call results
	callResult, err := mc.packer.GetCallResult(tryAggregate, ret)
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
