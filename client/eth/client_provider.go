package eth

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type ChainProvider interface {
	Reader
	Writer
	ConnectionPool
}

type ChainProviderImpl struct {
	ConnectionPool
}

func NewChainProviderImpl(pool ConnectionPool) (ChainProvider, error) {
	return &ChainProviderImpl{pool}, nil
}

// ==================================================================
// Implementations of Reader and Writer
// ==================================================================

func (c *ChainProviderImpl) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.GetBlockByNumber(ctx, number)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) GetReceipts(ctx context.Context, txs types.Transactions) (types.Receipts, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.GetReceipts(ctx, txs)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) GetReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.TransactionReceipt(ctx, txHash)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) SubscribeNewHead(ctx context.Context) (chan *types.Header, ethereum.Subscription, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.SubscribeNewHead(ctx)
	}
	return nil, nil, ErrClientNotFound
}

func (c *ChainProviderImpl) BlockNumber(ctx context.Context) (uint64, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.BlockNumber(ctx)
	}
	return 0, ErrClientNotFound
}

func (c *ChainProviderImpl) ChainID(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.ChainID(ctx)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.TransactionReceipt(ctx, txHash)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.GetBalance(ctx, address)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.CodeAt(ctx, account, blockNumber)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.EstimateGas(ctx, msg)
	}
	return 0, ErrClientNotFound
}

func (c *ChainProviderImpl) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.FilterLogs(ctx, q)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.HeaderByNumber(ctx, number)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.PendingCodeAt(ctx, account)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.PendingNonceAt(ctx, account)
	}
	return 0, ErrClientNotFound
}

func (c *ChainProviderImpl) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.SendTransaction(ctx, tx)
	}
	return ErrClientNotFound
}

func (c *ChainProviderImpl) SubscribeFilterLogs(
	ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log,
) (ethereum.Subscription, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.SubscribeFilterLogs(ctx, q, ch)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.SuggestGasPrice(ctx)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) CallContract(
	ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int,
) ([]byte, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.CallContract(ctx, msg, blockNumber)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetAnyChainClient(); ok {
		return client.SuggestGasTipCap(ctx)
	}
	return nil, ErrClientNotFound
}
