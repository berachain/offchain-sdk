package eth

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type ChainProvider interface {
	Reader
	Writer
	GetChainClient(string) (Client, bool)
	GetAnyChainClient() (Client, bool)
	AddChainClient(Config) bool
	RemoveChainClient(string) error
}

type ChainProviderImpl struct {
	cache  *lru.Cache[string, Client]
	mutex  sync.Mutex
	config ChainProviderConfig
}

type ChainProviderConfig struct {
	cacheSize      uint
	defaultTimeout time.Duration
}

func NewChainProviderImpl(cfg ChainProviderConfig) (ChainProvider, error) {
	cache, err := lru.NewWithEvict[string, Client](int(cfg.cacheSize), func(_ string, v Client) {
		defer v.Close()
		// The timeout is added so that any in progress requests have a chance to complete before we close.
		time.Sleep(cfg.defaultTimeout)
	})
	if err != nil {
		return nil, err
	}
	return &ChainProviderImpl{
		cache:  cache,
		config: cfg,
	}, nil
}

// GetChainClient returns a chain client from the cache.
// If clientAddr isn't specified, it returns any client from the cache.
func (c *ChainProviderImpl) GetChainClient(clientAddr string) (Client, bool) {
	// If a clientAddr is not specified, return any from the pool.
	if clientAddr == "" {
		return c.GetAnyChainClient()
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.cache.Get(clientAddr)
}

func (c *ChainProviderImpl) GetAnyChainClient() (Client, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, client, ok := c.cache.GetOldest()
	return client, ok
}

func (c *ChainProviderImpl) AddChainClient(cfg Config) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// If replacing, be sure to close old client first.
	// The LRU cache's eviction policy is not triggered on value updates/replacements.
	if c.cache.Contains(cfg.EthHTTPURL) {
		err := c.removeClient(cfg.EthHTTPURL)
		if err != nil {
			return false
		}
	}
	client := NewClient(&cfg)
	return c.cache.Add(cfg.EthHTTPURL, client)
}

func (c *ChainProviderImpl) RemoveChainClient(clientAddr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.removeClient(clientAddr)
}

func (c *ChainProviderImpl) removeClient(clientAddr string) error {
	client, ok := c.cache.Get(clientAddr)
	if !ok {
		return fmt.Errorf("could not get client for: %s", clientAddr)
	}
	client.Close()
	return nil
}

// ==================================================================
// Implementations of Reader and Writer
// ==================================================================

func (c *ChainProviderImpl) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.GetBlockByNumber(ctx, number)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) GetReceipts(ctx context.Context, txs types.Transactions) (types.Receipts, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.GetReceipts(ctx, txs)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) GetReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.TransactionReceipt(ctx, txHash)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) SubscribeNewHead(ctx context.Context) (chan *types.Header, ethereum.Subscription, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.SubscribeNewHead(ctx)
	}
	return nil, nil, ErrClientNotFound
}

func (c *ChainProviderImpl) BlockNumber(ctx context.Context) (uint64, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.BlockNumber(ctx)
	}
	return 0, ErrClientNotFound
}

func (c *ChainProviderImpl) ChainID(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.ChainID(ctx)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.TransactionReceipt(ctx, txHash)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.GetBalance(ctx, address)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.CodeAt(ctx, account, blockNumber)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.EstimateGas(ctx, msg)
	}
	return 0, ErrClientNotFound
}

func (c *ChainProviderImpl) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.FilterLogs(ctx, q)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.HeaderByNumber(ctx, number)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.PendingCodeAt(ctx, account)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.PendingNonceAt(ctx, account)
	}
	return 0, ErrClientNotFound
}

func (c *ChainProviderImpl) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if client, ok := c.GetChainClient(""); ok {
		return client.SendTransaction(ctx, tx)
	}
	return ErrClientNotFound
}

func (c *ChainProviderImpl) SubscribeFilterLogs(
	ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log,
) (ethereum.Subscription, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.SubscribeFilterLogs(ctx, q, ch)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.SuggestGasPrice(ctx)
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) CallContract(
	ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int,
) ([]byte, error) {
	if client, ok := c.GetChainClient(""); ok {
		return client.CallContract(ctx, msg, blockNumber)
	}
	return nil, ErrClientNotFound
}
