package baseapp

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/log"
	sdk "github.com/berachain/offchain-sdk/types"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// contextFacotry is used to produce new sdk.Contexts.
type contextFactory struct {
	connPool eth.Client
	logger   log.Logger
	db       ethdb.KeyValueStore
}

// NewContextFactory creates a new context from a given context.Context.
func (cf *contextFactory) NewSDKContext(ctx context.Context) *sdk.Context {
	return sdk.NewContext(
		ctx,
		cf.connPool,
		cf.logger,
		cf.db,
	)
}
