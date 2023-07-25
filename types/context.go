package types

import (
	"context"

	"github.com/berachain/offchain-sdk/log"
	"github.com/ethereum/go-ethereum/ethdb"
)

type Context struct {
	context.Context
	chain  Chain
	logger log.Logger
	db     ethdb.KeyValueStore
}

// UnwrapSdkContext unwraps the sdk context.
func UnwrapSdkContext(ctx context.Context) Context {
	if sdkCtx, ok := ctx.(Context); ok {
		return sdkCtx
	}
	panic("context is not sdk context")
}

func NewContext(ctx context.Context, chain Chain, logger log.Logger, db ethdb.KeyValueStore) *Context {
	return &Context{
		Context: ctx,
		chain:   chain,
		logger:  logger,
		db:      db,
	}
}

func (c *Context) Chain() Chain {
	return c.chain
}

func (c *Context) Logger() log.Logger {
	return c.logger
}

func (c *Context) DB() ethdb.KeyValueStore {
	return c.db
}
