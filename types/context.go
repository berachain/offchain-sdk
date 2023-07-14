package types

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/ethereum/go-ethereum/core/types"
)

type Chain interface {
	ChainReader
	ChainWriter
}

type ChainWriter interface{}

type ChainReader interface {
	CurrentBlock() (*types.Block, error)
	GetBlockByNumber(number uint64) (*types.Block, error)
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

type Context struct {
	context.Context
	chain Chain
	// chain Chain
}

func NewContext(ctx context.Context, chain Chain) *Context {
	return &Context{
		Context: ctx,
		chain:   chain,
	}
}

func (c *Context) Chain() Chain {
	return c.chain
}
