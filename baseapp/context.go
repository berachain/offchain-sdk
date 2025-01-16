package baseapp

import (
	"context"

	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/berachain/offchain-sdk/v2/log"
	"github.com/berachain/offchain-sdk/v2/telemetry"
	sdk "github.com/berachain/offchain-sdk/v2/types"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// contextFacotry is used to produce new sdk.Contexts.
type contextFactory struct {
	connPool eth.Client
	logger   log.Logger
	db       ethdb.KeyValueStore
	metrics  telemetry.Metrics
}

// NewContextFactory creates a new context from a given context.Context.
func (cf *contextFactory) NewSDKContext(ctx context.Context) *sdk.Context {
	return sdk.NewContext(
		ctx,
		cf.connPool,
		cf.logger,
		cf.db,
		cf.metrics,
	)
}
