package eth

import (
	"context"

	"github.com/ethereum/go-ethereum"
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

func (c *ContextualClient) SendTransaction(tx *types.Transaction) error {
	return c.client.SendTransaction(c.ctx, tx)
}
