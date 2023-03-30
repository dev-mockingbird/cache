package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type memoryCacheItem struct {
	start time.Time
	ttl   time.Duration
	value []byte
}

type memoryCache struct {
	data     map[string]*memoryCacheItem
	datalock sync.RWMutex
	sctrl    chan struct{}
}

func MemoryCache() Cache {
	m := &memoryCache{sctrl: make(chan struct{}), data: make(map[string]*memoryCacheItem)}
	m.cleanup()
	return m
}

func (c *memoryCache) Put(ctx context.Context, key string, val interface{}, ttl ...time.Duration) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.datalock.Lock()
	defer c.datalock.Unlock()
	c.data[key] = &memoryCacheItem{
		value: bs,
		start: time.Now(),
		ttl:   c.ttl(ttl),
	}
	return nil
}

func (c *memoryCache) cleanup() {
	go func() {
		ticker := time.Tick(time.Millisecond * 500)
		for {
			select {
			case <-ticker:
				go func() {
					now := time.Now()
					c.datalock.Lock()
					defer c.datalock.Unlock()
					for k, item := range c.data {
						if item.ttl == Permenent {
							continue
						}
						if now.Sub(item.start) > item.ttl {
							delete(c.data, k)
						}
					}
				}()
			case <-c.sctrl:
				return
			}
		}
	}()
}

func (c *memoryCache) Get(ctx context.Context, key string, val interface{}) error {
	c.datalock.RLock()
	defer c.datalock.RUnlock()
	if item, ok := c.data[key]; ok {
		return json.Unmarshal(item.value, val)
	}
	return ErrCacheNotFound
}

func (c *memoryCache) TTL(ctx context.Context, key string, ttl *time.Duration) error {
	c.datalock.RLock()
	defer c.datalock.RLock()
	if item, ok := c.data[key]; ok {
		*ttl = item.ttl - time.Now().Sub(item.start)
		return nil
	}
	return ErrCacheNotFound
}

func (c *memoryCache) Del(ctx context.Context, keys ...string) error {
	ch := make(chan struct{})
	go func() {
		c.datalock.Lock()
		defer c.datalock.Unlock()
		for _, k := range keys {
			delete(c.data, k)
		}
		ch <- struct{}{}
	}()
	<-ch
	return nil
}

func (c *memoryCache) Close() error {
	c.sctrl <- struct{}{}
	close(c.sctrl)
	return nil
}

func (c *memoryCache) ttl(ttl []time.Duration) time.Duration {
	if len(ttl) == 0 {
		return Permenent
	}
	return ttl[0]
}
