package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedis_Get(t *testing.T) {
	cache := Redis(redis.NewClient(&redis.Options{}))
	if err := cache.Get(context.Background(), "xxxxxxxxxxxxxxx", nil); err != nil {
		if !errors.Is(err, ErrCacheNotFound) {
			t.Fatal(err)
		}
	}
}
