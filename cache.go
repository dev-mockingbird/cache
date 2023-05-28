package cache

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrCacheNotFound, if the key not exists, this err should be return by Get and TTL
	ErrCacheNotFound = errors.New("cache not found")
)

// Cache key value cache abstraction
type Cache interface {
	// Put key with value and ttl
	Put(ctx context.Context, key string, val interface{}, ttl ...time.Duration) error
	// Get get value coresponding passed key
	Get(ctx context.Context, key string, val interface{}) error
	// TTL get the rest ttl coresponding key
	TTL(ctx context.Context, key string, ttl *time.Duration) error
	// Del remove a bunch of key and value
	Del(ctx context.Context, keys ...string) error
}
