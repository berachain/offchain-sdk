package types

import (
	"context"

	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/berachain/offchain-sdk/v2/log"
	"github.com/berachain/offchain-sdk/v2/telemetry"

	"github.com/ethereum/go-ethereum/ethdb"
)

type CancellableContext interface {
	Cancel()
}

type Context struct {
	context.Context
	chain   eth.Client
	logger  log.Logger
	db      ethdb.KeyValueStore
	metrics telemetry.Metrics
}

// UnwrapContext unwraps the sdk context.
func UnwrapContext(ctx context.Context) *Context {
	if sdkCtx, ok := ctx.(*Context); ok {
		return sdkCtx
	}

	panic("context is not sdk context")
}

func NewContext(
	ctx context.Context,
	ethClient eth.Client,
	logger log.Logger,
	db ethdb.KeyValueStore,
	metrics telemetry.Metrics,
) *Context {
	return &Context{
		Context: ctx,
		chain:   ethClient,
		logger:  logger,
		db:      db,
		metrics: metrics,
	}
}

func (c *Context) Chain() eth.Client {
	return c.chain
}

func (c *Context) Logger() log.Logger {
	return c.logger
}

func (c *Context) DB() ethdb.KeyValueStore {
	return c.db
}

func (c *Context) Metrics() telemetry.Metrics {
	return c.metrics
}
