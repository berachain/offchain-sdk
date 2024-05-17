package store

import (
	"context"
	"time"
)

type Store interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (interface{}, time.Duration, error)
	Increment(ctx context.Context, key string) (int64, time.Duration, error)
	Remove(ctx context.Context, key string) error
}
