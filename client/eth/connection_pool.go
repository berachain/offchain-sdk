package eth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/log"
	lru "github.com/hashicorp/golang-lru/v2"
)

type ConnectionPool interface {
	GetHTTP() (*HealthCheckedClient, bool)
	GetWS() (*HealthCheckedClient, bool)
	RemoveChainClient(string) error
	Close() error
	Dial(string) error
	DialContext(context.Context, string) error
}

type ConnectionPoolImpl struct {
	cache   *lru.Cache[string, *HealthCheckedClient]
	wsCache *lru.Cache[string, *HealthCheckedClient]
	mutex   sync.Mutex
	config  ConnectionPoolConfig
	logger  log.Logger
}

type ConnectionPoolConfig struct {
	EthHTTPURLs    []string
	EthWSURLs      []string
	DefaultTimeout time.Duration
}

func DefaultConnectPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		EthHTTPURLs: []string{"http://localhost:8545"},
		EthWSURLs:   []string{"ws://localhost:8546"},
	}
}

func NewConnectionPoolImpl(cfg ConnectionPoolConfig, logger log.Logger) (ConnectionPool, error) {
	cache, err := lru.NewWithEvict[string, *HealthCheckedClient](
		len(cfg.EthHTTPURLs), func(_ string, v *HealthCheckedClient) {
			defer v.Close()
			// The timeout is added so that any in progress
			// requests have a chance to complete before we close.
			time.Sleep(cfg.DefaultTimeout)
		})
	if err != nil {
		return nil, err
	}
	wsCache, err := lru.NewWithEvict[string, *HealthCheckedClient](
		len(cfg.EthHTTPURLs), func(_ string, v *HealthCheckedClient) {
			defer v.Close()
			// The timeout is added so that any in progress
			// requests have a chance to complete before we close.
			time.Sleep(cfg.DefaultTimeout)
		})
	if err != nil {
		return nil, err
	}

	return &ConnectionPoolImpl{
		cache:   cache,
		wsCache: wsCache,
		config:  cfg,
		logger:  logger,
	}, nil
}

func (c *ConnectionPoolImpl) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, client := range c.cache.Keys() {
		if err := c.removeClient(client); err != nil {
			return err
		}
	}
	return nil
}

func (c *ConnectionPoolImpl) Dial(string) error {
	return c.DialContext(context.Background(), "")
}

func (c *ConnectionPoolImpl) DialContext(ctx context.Context, _ string) error {
	for _, url := range c.config.EthHTTPURLs {
		client := NewHealthCheckedClient(c.logger)
		if err := client.DialContext(ctx, url); err != nil {
			return err
		}
		c.cache.Add(url, client)
	}
	for _, url := range c.config.EthWSURLs {
		client := NewHealthCheckedClient(c.logger)
		if err := client.DialContext(ctx, url); err != nil {
			return err
		}
		c.wsCache.Add(url, client)
	}
	return nil
}

func (c *ConnectionPoolImpl) GetHTTP() (*HealthCheckedClient, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
retry:
	_, client, ok := c.cache.GetOldest()
	if !client.Healthy() {
		goto retry
	}
	return client, ok
}

func (c *ConnectionPoolImpl) GetWS() (*HealthCheckedClient, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
retry:
	_, client, ok := c.wsCache.GetOldest()
	if !client.Healthy() {
		goto retry
	}
	return client, ok
}

func (c *ConnectionPoolImpl) RemoveChainClient(clientAddr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.removeClient(clientAddr)
}

func (c *ConnectionPoolImpl) removeClient(clientAddr string) error {
	client, ok := c.cache.Get(clientAddr)
	if !ok {
		return fmt.Errorf("could not get client for: %s", clientAddr)
	}
	client.Close()
	return nil
}
