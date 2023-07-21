package types

import (
	"context"

	"github.com/berachain/offchain-sdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

type Chain interface {
	ChainReader
	ChainWriter
	ChainSubscriber
}

type ChainWriter interface{}

type ChainReader interface {
	CurrentBlock() (*types.Block, error)
	GetBlockByNumber(number uint64) (*types.Block, error)
}

type ChainSubscriber interface {
	SubscribeFilterLogs(q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
}

type Context struct {
	context.Context
	chain  Chain
	logger log.Logger
	ms     MultiStore
}

// UnwrapSdkContext unwraps the sdk context.
func UnwrapSdkContext(ctx context.Context) Context {
	if sdkCtx, ok := ctx.(Context); ok {
		return sdkCtx
	}
	panic("context is not sdk context")
}

func NewContext(ctx context.Context, chain Chain, logger log.Logger, ms MultiStore) *Context {
	return &Context{
		Context: ctx,
		chain:   chain,
		logger:  logger,
		ms:      ms,
	}
}

func (c *Context) Chain() Chain {
	return c.chain
}

func (c *Context) Logger() log.Logger {
	return c.logger
}

func (c *Context) GetStore(name string) KVStore {
	return c.ms.GetKVStore(name)
}
