package eth

import (
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type ConnectionPool interface {
	GetChainClient(string) (Client, bool)
	GetAnyChainClient() (Client, bool)
	AddChainClient(Config) bool
	RemoveChainClient(string) error
}

type ConnectionPoolImpl struct {
	cache  *lru.Cache[string, Client]
	mutex  sync.Mutex
	config ConnectionPoolConfig
}

type ConnectionPoolConfig struct {
	cacheSize      uint
	defaultTimeout time.Duration
}

func NewConnectionPoolImpl(cfg ConnectionPoolConfig) (ConnectionPool, error) {
	cache, err := lru.NewWithEvict[string, Client](int(cfg.cacheSize), func(_ string, v Client) {
		defer v.Close()
		// The timeout is added so that any in progress requests have a chance to complete before we close.
		time.Sleep(cfg.defaultTimeout)
	})
	if err != nil {
		return nil, err
	}
	return &ConnectionPoolImpl{
		cache:  cache,
		config: cfg,
	}, nil
}

// GetChainClient returns a chain client from the cache.
// If clientAddr isn't specified, it returns any client from the cache.
func (c *ConnectionPoolImpl) GetChainClient(clientAddr string) (Client, bool) {
	// If a clientAddr is not specified, return any from the pool.
	if clientAddr == "" {
		return c.GetAnyChainClient()
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.cache.Get(clientAddr)
}

func (c *ConnectionPoolImpl) GetAnyChainClient() (Client, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, client, ok := c.cache.GetOldest()
	return client, ok
}

func (c *ConnectionPoolImpl) AddChainClient(cfg Config) bool {
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
