package batcher_test

import (
	"math/big"
	"os"
	"testing"

	"github.com/berachain/offchain-sdk/contracts/bindings"
	"github.com/berachain/offchain-sdk/core/transactor/factory/batcher"
	"github.com/berachain/offchain-sdk/core/transactor/types"

	"github.com/ethereum/go-ethereum/common"
)

// TestPayableMulticall demonstrates how to use the PayableMulticall contract to batch multiple
// calls to a specific contract on a Ethereum blockchain.
//
// NOTE: the following must be set up for this test to run:
//  1. This test will only run if the chain is available at env var `ETH_RPC_URL`.
//  2. The PayableMulticallable contract must be deployed at env var `PAYABLE_MULTICALL_ADDR`
//     (example contract can be found at offchain-sdk/contracts/src/PayableMulticall.sol).
//  3. Requires a EOA wallet with some ETH to "pay value" at env variable `WALLET_ADDR`!
func TestPayableMulticall(t *testing.T) {
	// setup inputs, eth client, and multicaller
	walletAddr := common.HexToAddress(os.Getenv("WALLET_ADDR"))
	if walletAddr == empty {
		t.Skipf("Skipping test: no private key provided")
	}
	payableMulticallAddr := common.HexToAddress(os.Getenv("PAYABLE_MULTICALL_ADDR"))
	if payableMulticallAddr == empty {
		t.Skipf("Skipping test: no payable multicall address provided")
	}
	sCtx := setUp(t)
	multicaller := batcher.NewPayableMulticall(payableMulticallAddr)

	// set up some test calls to make
	pmcPacker := types.Packer{MetaData: bindings.PayableMulticallMetaData}
	call1, err := pmcPacker.CreateRequest(
		"", payableMulticallAddr, big.NewInt(1), nil, nil, 0, "incNumber",
	)
	if err != nil {
		t.Fatal(err)
	}
	// this call will revert bc of a value of 0, but the batch should still succeed
	call2, err := pmcPacker.CreateRequest(
		"", payableMulticallAddr, big.NewInt(0), nil, nil, 0, "incNumber",
	)
	if err != nil {
		t.Fatal(err)
	}
	call3, err := pmcPacker.CreateRequest(
		"", payableMulticallAddr, big.NewInt(3), nil, nil, 0, "incNumber",
	)
	if err != nil {
		t.Fatal(err)
	}

	// batch and send the calls to the chain
	resp, err := multicaller.BatchCallRequests(
		sCtx, walletAddr, call1.CallMsg, call2.CallMsg, call3.CallMsg,
	)
	if err != nil {
		t.Fatal(err)
	}
	responses, ok := resp.([][]byte)
	if !ok {
		t.Fatalf("expected [][]byte, got %T", resp)
	}

	t.Log("responses", responses)
}
