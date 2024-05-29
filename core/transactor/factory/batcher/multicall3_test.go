package batcher_test

import (
	"math/big"
	"os"
	"testing"

	"github.com/berachain/offchain-sdk/contracts/bindings"
	"github.com/berachain/offchain-sdk/core/transactor/factory/batcher"
	"github.com/berachain/offchain-sdk/core/transactor/types"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
)

// TestMulticall3 demonstrates how to use the multicall3 contract to batch multiple calls to other
// contracts on a Ethereum blockchain.
//
// NOTE: the following must be set up for this test to run:
//  1. This test will only run if the chain is available at env var `ETH_RPC_URL`.
//  2. The Multicall3 contract must be deployed at env var `MULTICALL3_ADDR`.
//  3. Any ERC20 contract must be deployed at env var `ERC20_ADDR`.
func TestMulticall3(t *testing.T) {
	// setup inputs, eth client, and multicaller
	multicall3Addr := common.HexToAddress(os.Getenv("MULTICALL3_ADDR"))
	if multicall3Addr == empty {
		t.Skipf("Skipping test: no multicall3 address provided")
	}
	erc20Addr := common.HexToAddress(os.Getenv("ERC20_ADDR"))
	if erc20Addr == empty {
		t.Skipf("Skipping test: no erc20 address provided")
	}
	sCtx := setUp(t)
	multicaller := batcher.NewMulticall3(multicall3Addr)

	// set up some test calls to make
	mc3Packer := types.Packer{MetaData: bindings.Multicall3MetaData}
	call1, err := mc3Packer.CreateRequest("", multicall3Addr, nil, nil, nil, 0, "getBlockNumber")
	if err != nil {
		t.Fatal(err)
	}
	erc20Packer := types.Packer{MetaData: bindings.IERC20MetaData}
	call2, err := erc20Packer.CreateRequest("", erc20Addr, nil, nil, nil, 0, "balanceOf", empty)
	if err != nil {
		t.Fatal(err)
	}

	// batch and send the calls to the chain
	resp, err := multicaller.BatchCallRequests(sCtx, empty, false, call1.CallMsg, call2.CallMsg)
	if err != nil {
		t.Fatal(err)
	}
	responses, ok := resp.([]bindings.Multicall3Result)
	if !ok {
		t.Fatalf("expected []bindings.Multicall3Result, got %T", resp)
	}
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	assert.True(t, responses[0].Success)
	assert.True(t, responses[1].Success)

	// unpack the first response
	ret1, err := mc3Packer.GetCallResult("getBlockNumber", responses[0].ReturnData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(ret1))
	assert.Less(t, uint64(0), ret1[0].(*big.Int).Uint64())

	// unpack the second response
	ret2, err := erc20Packer.GetCallResult("balanceOf", responses[1].ReturnData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(ret2))
	assert.Equal(t, uint64(0), ret2[0].(*big.Int).Uint64())
}
