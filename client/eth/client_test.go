package eth_test

import (
	"context"
	"testing"
	"time"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/config/env"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// This file tests the methods on the extended eth client.
//
// Note that the following must be set up for these tests to run:
//  - the chain RPC is available at env var `ETH_RPC_URL`
//  - OR the chain WS RPC is available at env var `ETH_RPC_URL_WS`
//  - the RPC endpoint must return within 5 seconds or tests will timeout
//  - [Optional for `txPoolContentFrom`] a wallet to query the mempool for at env var `ETH_ADDR`

const (
	TestModeHTTP int = iota
	TestModeWS
	TestModeEither
)

// setupClientTest loads environment variables and performs any necessary test setup.
func setupClientTest(t *testing.T) {
	t.Helper()
	err := env.Load()
	assert.NoError(t, err)
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL` or `ETH_RPC_URL_WS`.
func setUp(testMode int, t *testing.T) (*eth.ExtendedEthClient, error) {
	setupClientTest(t)
	rpcTimeout := 5 * time.Second
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()

	var ethRPC string
	switch testMode {
	case TestModeWS:
		ethRPC = env.GetEthWSURL()
	case TestModeHTTP:
		ethRPC = env.GetEthRPCURL()
	case TestModeEither:
		if ethRPC = env.GetEthWSURL(); ethRPC == "" {
			ethRPC = env.GetEthRPCURL()
		}
	default:
		panic("invalid test mode")
	}
	if ethRPC == "" {
		t.Skipf("Skipping test: no eth rpc url provided")
	}

	ethClient, err := ethclient.DialContext(ctxWithTimeout, ethRPC)
	if err != nil {
		return nil, err
	}

	eec := eth.NewExtendedEthClient(ethClient, rpcTimeout)
	return eec, nil
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL` or `ETH_RPC_URL_WS`.
func TestClose(t *testing.T) {
	eec, err := setUp(TestModeEither, t)
	assert.NoError(t, err)

	err = eec.Close()
	assert.NoError(t, err)
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL` or `ETH_RPC_URL_WS`.
func TestHealth(t *testing.T) {
	eec, err := setUp(TestModeEither, t)
	assert.NoError(t, err)

	health := eec.Health()
	assert.True(t, health)
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL`.
func TestGetReceipts(t *testing.T) {
	eec, err := setUp(TestModeHTTP, t)
	assert.NoError(t, err)

	ctx := context.Background()
	txs := ethcoretypes.Transactions{}

	receipts, err := eec.GetReceipts(ctx, txs)
	assert.NoError(t, err)
	assert.Empty(t, receipts)
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL_WS`.
func TestSubscribeNewHead(t *testing.T) {
	eec, err := setUp(TestModeWS, t)
	assert.NoError(t, err)

	ctx := context.Background()

	ch, sub, err := eec.SubscribeNewHead(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ch)
	assert.NotNil(t, sub)

	assert.NotPanics(t, func() { sub.Unsubscribe() })
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL_WS`.
func TestSubscribeFilterLogs(t *testing.T) {
	eec, err := setUp(TestModeWS, t)
	assert.NoError(t, err)

	ctx := context.Background()
	query := ethereum.FilterQuery{}
	ch := make(chan ethcoretypes.Log)

	sub, err := eec.SubscribeFilterLogs(ctx, query, ch)
	assert.NoError(t, err)
	assert.NotNil(t, sub)

	assert.NotPanics(t, func() { sub.Unsubscribe() })
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL` AND a wallet to query the
// txpool for at `ETH_ADDR`.
func TestTxPoolContentFrom(t *testing.T) {
	eec, err := setUp(TestModeHTTP, t)
	assert.NoError(t, err)

	ctx := context.Background()
	addrStr := env.GetAddressToListen()
	if addrStr == "" {
		t.Skipf("Skipping test: no eth address provided")
	}
	address := common.HexToAddress(addrStr)

	result, err := eec.TxPoolContentFrom(ctx, address)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	t.Log("result", result)
}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL`.
func TestTxPoolInspect(t *testing.T) {
	eec, err := setUp(TestModeHTTP, t)
	assert.NoError(t, err)

	ctx := context.Background()

	result, err := eec.TxPoolInspect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
