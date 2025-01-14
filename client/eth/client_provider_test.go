package eth_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/berachain/offchain-sdk/v2/client/eth/mocks"
	tmocks "github.com/berachain/offchain-sdk/v2/telemetry/mocks"
	"github.com/stretchr/testify/mock"
)

func mockedChainProvider() (*mocks.Client, eth.Client) {
	mockedCP := new(mocks.ConnectionPool)
	chainProvider := eth.NewChainProviderImpl(mockedCP, eth.ConnectionPoolConfig{
		DefaultTimeout: 5,
	})
	mockedClient := new(mocks.Client)
	mockedCP.On("GetHTTP").Return(mockedClient, true)
	mockedClient.On("ClientID").Return("test")
	return mockedClient, chainProvider
}

// TestMetricsEmitOnRpcCall tests that metrics are emitted when an RPC call is made.
func TestMetricsEmitOnRpcCall(t *testing.T) {
	mockedRPC, chainProvider := mockedChainProvider()
	mockedMetrics := new(tmocks.Metrics)
	cpImpl, ok := chainProvider.(*eth.ChainProviderImpl)
	if !ok {
		t.Fatal("chainProvider is not an instance of ChainProviderImpl")
	}
	cpImpl.EnableMetrics(mockedMetrics)

	mockedRPC.On("BlockByNumber", mock.Anything, mock.Anything).Return(nil, nil)

	expectedArg1 := "rpc_id:test"
	expectedArg2 := "method:eth_getBlockByNumber"

	mockedMetrics.On(
		"IncMonotonic",
		mock.Anything,
		expectedArg1,
		expectedArg2,
	).Once()

	mockedMetrics.On(
		"Time",
		mock.Anything,
		mock.Anything,
		expectedArg1,
		expectedArg2,
	).Once()

	_, err := chainProvider.BlockByNumber(context.Background(), big.NewInt(1))
	mockedMetrics.AssertExpectations(t)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
