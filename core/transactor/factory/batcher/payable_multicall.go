package batcher

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

const multicall = `multicall`

// Corresponding to the PayableMulticall contract in contracts/lib/transient-goodies/src
// (https://github.com/berachain/transient-goodies/blob/try-aggregate/src/PayableMulticallable.sol)
type PayableMulticall struct {
	contractAddress common.Address
	packer          *types.Packer
}

// NewPayableMulticall creates a new PayableMulticall instance.
func NewPayableMulticall(address common.Address) *PayableMulticall {
	return &PayableMulticall{
		contractAddress: address,
		packer:          &types.Packer{MetaData: bindings.PayableMulticallableMetaData},
	}
}

// BatchRequests creates a batched transaction request for the given call requests.
func (mc *PayableMulticall) BatchRequests(callReqs ...*ethereum.CallMsg) *types.Request {
	var (
		calls       = make([][]byte, len(callReqs))
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

		// set the calldata.
		calls[i] = callReq.Data
	}

	txRequest, _ := mc.packer.CreateRequest(
		"", mc.contractAddress, totalValue, gasTipCap, gasFeeCap, gasLimit,
		multicall, false, calls,
	)
	return txRequest
}

// BatchCallRequests uses the PayableMulticall contract to create a batched call request for the
// given call messages and return the batched call result data for each call, as a `[][]byte`.
func (mc *PayableMulticall) BatchCallRequests(
	ctx context.Context, from common.Address, callReqs ...*ethereum.CallMsg,
) (any, error) {
	sCtx := sdk.UnwrapContext(ctx)

	// get the batched tx (call) requests
	batchedCall := mc.BatchRequests(callReqs...)
	batchedCall.From = from

	// call the multicall3 contract with the batched call request
	ret, err := sCtx.Chain().CallContract(ctx, *batchedCall.CallMsg, nil)
	if err != nil {
		if _, reason, ok := strings.Cut(err.Error(), executionReverted); ok {
			sCtx.Logger().Warn("execution reverted for payable multicall", "reason", reason)
		} else {
			sCtx.Logger().Error("failed to call contract", "err", err)
		}
		return nil, err
	}

	// unpack the return data into call results
	callResult, err := mc.packer.GetCallResult(multicall, ret)
	if err != nil {
		sCtx.Logger().Error("failed to unpack call response", "err", err)
		return nil, err
	}
	if len(callResult) != 1 {
		err = fmt.Errorf("expected 1 list of [][]byte, got %d", len(callResult))
		sCtx.Logger().Error("failed to unpack call response", "err", err)
		return nil, err
	}
	callResults, ok := callResult[0].([][]byte)
	if !ok {
		err = errors.New("expected return type as list of bytes[]")
		sCtx.Logger().Error("failed to unpack call response", "err", err)
		return nil, err
	}

	return callResults, nil
}
