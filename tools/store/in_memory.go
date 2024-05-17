package store

import (
	"context"
	"errors"
	"time"

	"github.com/jellydator/ttlcache/v2"
)

type InMemoryStore struct {
	cache *ttlcache.Cache
	ttl   time.Duration
}

func NewInMemoryStore(ttl time.Duration) Store {
	cache := ttlcache.NewCache()
	cache.SkipTTLExtensionOnHit(true)
	return &InMemoryStore{
		cache: cache,
		ttl:   ttl,
	}
}

func (c *InMemoryStore) Set(_ context.Context, key string, value interface{}) error {
	return c.cache.SetWithTTL(key, value, c.ttl)
}

func (c *InMemoryStore) Increment(_ context.Context, key string) (int64, time.Duration, error) {
	item, exp, err := c.cache.GetWithTTL(key)
	if errors.Is(err, ttlcache.ErrNotFound) {
		err = c.cache.SetWithTTL(key, int64(1), c.ttl)
		return 1, c.ttl, err
	} else if err != nil {
		return 0, 0, err
	}
	count := item.(int64) + 1

	err = c.cache.SetWithTTL(key, count, exp)
	if err != nil {
		return 0, 0, err
	}
	return count, exp, nil
}

func (c *InMemoryStore) Get(_ context.Context, key string) (interface{}, time.Duration, error) {
	return c.cache.GetWithTTL(key)
}

func (c *InMemoryStore) Remove(_ context.Context, key string) error {
	return c.cache.Remove(key)
}
