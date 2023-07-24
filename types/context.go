package types

import (
	"context"

	"github.com/berachain/offchain-sdk/log"
)

type Context struct {
	context.Context
	chain  Chain
	logger log.Logger
}

// UnwrapSdkContext unwraps the sdk context.
func UnwrapSdkContext(ctx context.Context) Context {
	if sdkCtx, ok := ctx.(Context); ok {
		return sdkCtx
	}
	panic("context is not sdk context")
}

func NewContext(ctx context.Context, chain Chain, logger log.Logger) *Context {
	return &Context{
		Context: ctx,
		chain:   chain,
		logger:  logger,
	}
}

func (c *Context) Chain() Chain {
	return c.chain
}

func (c *Context) Logger() log.Logger {
	return c.logger
}
