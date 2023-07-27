package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ContextualClient struct {
	client Client
	ctx    context.Context
}

func NewContextualClient(ctx context.Context, client Client) *ContextualClient {
	return &ContextualClient{
		client: client,
		ctx:    ctx,
	}
}

func (c *ContextualClient) CurrentBlock() (*types.Block, error) {
	x, err := c.client.BlockNumber(c.ctx)
	if err != nil {
		return nil, err
	}
	return c.client.GetBlockByNumber(c.ctx, x)
}

func (c *ContextualClient) GetBlockByNumber(number uint64) (*types.Block, error) {
	return c.client.GetBlockByNumber(c.ctx, number)
}

func (c *ContextualClient) SubscribeNewHead() (chan *types.Header,
	ethereum.Subscription, error) {
	return c.client.SubscribeNewHead(c.ctx)
}

func (c *ContextualClient) SubscribeFilterLogs(q ethereum.FilterQuery,
	ch chan<- types.Log) (ethereum.Subscription, error) {
	return c.client.SubscribeFilterLogs(c.ctx, q, ch)
}

func (c *ContextualClient) SendTransaction(_ context.Context, tx *types.Transaction) error {
	return c.client.SendTransaction(c.ctx, tx)
}

func (c *ContextualClient) CodeAt(_ context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return c.client.CodeAt(c.ctx, contract, blockNumber)
}

func (c *ContextualClient) CallContract(_ context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return c.client.CallContract(c.ctx, msg, blockNumber)
}

func (c *ContextualClient) PendingCodeAt(_ context.Context, account common.Address) ([]byte, error) {
	return c.client.PendingCodeAt(c.ctx, account)
}

func (c *ContextualClient) EstimateGas(_ context.Context, msg ethereum.CallMsg) (uint64, error) {
	return c.client.EstimateGas(c.ctx, msg)
}

func (c *ContextualClient) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	return c.client.HeaderByNumber(c.ctx, number)
}

func (c *ContextualClient) FilterLogs(_ context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	return c.client.FilterLogs(c.ctx, query)
}

func (c *ContextualClient) ChainID(_ context.Context) (*big.Int, error) {
	return c.client.ChainID(c.ctx)
}

func (c *ContextualClient) PendingNonceAt(_ context.Context, account common.Address) (uint64, error) {
	return c.client.PendingNonceAt(c.ctx, account)
}
