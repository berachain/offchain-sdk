package types

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
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

var _ Chain = (*ContextEthClient)(nil)

type ContextEthClient struct {
	eth.Client
	Ctx context.Context
}

func (c *ContextEthClient) CurrentBlock() (*types.Block, error) {
	x, err := c.Client.BlockNumber(c.Ctx)
	if err != nil {
		return nil, err
	}
	return c.Client.GetBlockByNumber(c.Ctx, x)
}

func (c *ContextEthClient) GetBlockByNumber(number uint64) (*types.Block, error) {
	return c.Client.GetBlockByNumber(c.Ctx, number)
}

func (c *ContextEthClient) SubscribeNewHead() (chan *types.Header, ethereum.Subscription, error) {
	return c.Client.SubscribeNewHead(c.Ctx)
}

func (c *ContextEthClient) SubscribeFilterLogs(q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return c.Client.SubscribeFilterLogs(c.Ctx, q, ch)
}

type Context struct {
	context.Context
	chain  Chain
	logger log.Logger
	// chain Chain
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
