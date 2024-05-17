package store

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	// Add more methods as needed
}

func NewRedisClient(addr string, clusterMode bool) RedisClient {
	if clusterMode {
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{addr},
		})
	}
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

type RedisStore struct {
	client RedisClient
	ttl    time.Duration
}

func NewRedisStore(ttl time.Duration, addr string, clusterMode bool) Store {
	client := NewRedisClient(addr, clusterMode)

	_, err := client.Ping(context.TODO()).Result()
	if err != nil {
		panic(err)
	}

	return &RedisStore{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisStore) Set(ctx context.Context, key string, value interface{}) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

func (c *RedisStore) Increment(ctx context.Context, key string) (int64, time.Duration, error) {
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	if count == 1 {
		err = c.client.Expire(ctx, key, c.ttl).Err()
		if err != nil {
			return 0, 0, err
		}
		return count, c.ttl, nil
	}
	expiration, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	return count, expiration, nil
}

func (c *RedisStore) Get(ctx context.Context, key string) (interface{}, time.Duration, error) {
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	expiration, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, 0, err
	}

	return result, expiration, nil
}

func (c *RedisStore) Remove(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
