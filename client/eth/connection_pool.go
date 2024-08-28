package eth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/log"
	lru "github.com/hashicorp/golang-lru/v2"
)

//go:generate mockery --name ConnectionPool
type ConnectionPool interface {
	GetHTTP() (Client, bool)
	GetWS() (Client, bool)
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

func NewConnectionPoolImpl(cfg ConnectionPoolConfig, logger log.Logger) (ConnectionPool, error) {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = defaultRPCTimeout
	}
	if cfg.HealthCheckInterval == 0 {
		cfg.HealthCheckInterval = defaultHealthCheckInterval
	}

	var (
		cache   *lru.Cache[string, *HealthCheckedClient]
		wsCache *lru.Cache[string, *HealthCheckedClient]
		err     error
	)

	// The LRU cache needs at least one URL provided for HTTP.
	if len(cfg.EthHTTPURLs) == 0 {
		return nil, fmt.Errorf("ConnectionPool: missing URL for HTTP clients")
	}

	cache, err = lru.NewWithEvict(
		len(cfg.EthHTTPURLs), func(_ string, v *HealthCheckedClient) {
			defer v.Close()
			// The timeout is added so that any in progress
			// requests have a chance to complete before we close.
			time.Sleep(cfg.DefaultTimeout)
		})
	if err != nil {
		return nil, err
	}

	if len(cfg.EthWSURLs) == 0 {
		logger.Warn("ConnectionPool: missing URL for WS clients")
	} else {
		wsCache, err = lru.NewWithEvict(
			len(cfg.EthWSURLs), func(_ string, v *HealthCheckedClient) {
				defer v.Close()
				// The timeout is added so that any in progress
				// requests have a chance to complete before we close.
				time.Sleep(cfg.DefaultTimeout)
			})
		if err != nil {
			return nil, err
		}
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

	if c.cache == nil {
		return nil
	}

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
	// NOTE: Check the cache for the HTTP URL is not needed because it
	// is guaranteed to be non-nil when a ConnectionPoolImpl is created.
	for _, url := range c.config.EthHTTPURLs {
		client := NewHealthCheckedClient(c.config.HealthCheckInterval, c.logger)
		if err := client.DialContextWithTimeout(ctx, url, c.config.DefaultTimeout); err != nil {
			return err
		}
		c.cache.Add(url, client)
	}

	// Check is needed because the WS URL is optional.
	if c.wsCache == nil {
		return nil
	}

	for _, url := range c.config.EthWSURLs {
		client := NewHealthCheckedClient(c.config.HealthCheckInterval, c.logger)
		if err := client.DialContextWithTimeout(ctx, url, c.config.DefaultTimeout); err != nil {
			return err
		}
		c.wsCache.Add(url, client)
	}
	return nil
}

// NOTE: this function assumes the cache is non-nil
// because it is guaranteed to be non-nil when a ConnectionPoolImpl is created.
func (c *ConnectionPoolImpl) GetHTTP() (Client, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.getClientFrom(c.cache)
}

func (c *ConnectionPoolImpl) GetWS() (Client, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Because the WS URL is optional, we need to check if it's nil.
	if c.wsCache == nil {
		return nil, false
	}
	return c.getClientFrom(c.wsCache)
}

// NOTE: this function assumes the lock is held and cache is non-nil.
func (c *ConnectionPoolImpl) getClientFrom(
	cache *lru.Cache[string, *HealthCheckedClient],
) (Client, bool) {
retry:
	_, client, ok := cache.GetOldest()
	if !client.Health() {
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
