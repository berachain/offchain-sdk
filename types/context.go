package types

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/log"
	"github.com/ethereum/go-ethereum/ethdb"
)

type Context struct {
	context.Context
	chain  eth.Client
	logger log.Logger
	db     ethdb.KeyValueStore
}

// UnwrapSdkContext unwraps the sdk context.
func UnwrapSdkContext(ctx context.Context) *Context {
	if sdkCtx, ok := ctx.(*Context); ok {
		return sdkCtx
	}

	panic("context is not sdk context")
}

func NewContext(ctx context.Context, ethClient eth.Client, logger log.Logger, db ethdb.KeyValueStore) *Context {
	return &Context{
		Context: ctx,
		chain:   ethClient,
		logger:  logger,
		db:      db,
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
