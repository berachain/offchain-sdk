package eth

import (
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type ClientPool interface {
	GetClient(string) (Client, bool)
	AddClient(Config, time.Duration) bool
	RemoveClient(string) error
}

type ClientPoolImpl struct {
	cache *lru.Cache[string, Client]
	cfg   ClientPoolConfig
	mutex sync.Mutex
}

type ClientPoolConfig struct {
	cacheSize      uint
	defaultTimeout time.Duration
}

func NewClientPool(cfg ClientPoolConfig) (ClientPool, error) {
	cache, err := lru.NewWithEvict[string, Client](int(cfg.cacheSize), func(_ string, v Client) {
		defer v.Close()
		// The timeout is added so that any in progress requests have a chance to complete before we close.
		time.Sleep(cfg.defaultTimeout)
	})
	if err != nil {
		return nil, err
	}
	return ClientPoolImpl{
		cache: cache,
		cfg:   cfg,
	}, nil
}

// GetClient is the getter helper function, to retrieve clients given a key.
func (c ClientPoolImpl) GetClient(address string) (Client, bool) {
	return c.cache.Get(address)
}

// AddClient adds a client to the connection pool, and if an eviction occurred, it returns true.
func (c ClientPoolImpl) AddClient(cfg Config, timeout time.Duration) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If replacing, be sure to close old client first.
	// The LRU cache's eviction policy is not triggered on value updates/replacements.
	if c.cache.Contains(cfg.EthHTTPURL) {
		c.removeClient(cfg.EthHTTPURL)
	}

	client := NewClient(&cfg)
	return c.cache.Add(cfg.EthHTTPURL, client)
}

// RemoveClient is a concurrency safe helper to remove a client from the pool.
func (c ClientPoolImpl) RemoveClient(address string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.removeClient(address)
}

// removeClient is not concurrency safe, and is a helper to remove a client from the pool.
func (c ClientPoolImpl) removeClient(address string) error {
	client, ok := c.cache.Get(address)
	if !ok {
		return fmt.Errorf("Could not get client for: %s", address)
	}
	client.Close()
	return nil
}
