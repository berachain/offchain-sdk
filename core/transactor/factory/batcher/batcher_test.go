package batcher_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/berachain/offchain-sdk/v2/log"
	sdk "github.com/berachain/offchain-sdk/v2/types"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
)

var empty = common.Address{}

// NOTE: requires Ethereum chain rpc url at env var `ETH_RPC_URL`.
func setUp(t *testing.T) *sdk.Context {
	ctx := context.Background()
	ethRPC := os.Getenv("ETH_RPC_URL")
	if ethRPC == "" {
		t.Skipf("Skipping test: no eth rpc url provided")
	}
	chain, err := ethclient.DialContext(ctx, ethRPC)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = chain.ChainID(ctx); err != nil {
		if assert.ErrorContains(t, err, "connection refused") {
			t.Skipf("Skipping test: %s", err)
		}
		t.Fatal(err)
	}
	return sdk.NewContext(
		ctx, eth.NewExtendedEthClient(chain, 5*time.Second),
		log.NewLogger(os.Stdout, "test-runner"), nil, nil,
	)
}
