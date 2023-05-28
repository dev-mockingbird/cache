package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

var globalCache map[string]*memoryCache
var globallock sync.RWMutex

func init() {
	globalCache = make(map[string]*memoryCache)
}

type memoryCacheItem struct {
	start time.Time
	ttl   time.Duration
	value []byte
}

type memoryCache struct {
	data     map[string]*memoryCacheItem
	datalock sync.RWMutex
	group    string
	sctrl    chan struct{}
}

func Memory(group string) Cache {
	globallock.RLock()
	if m, ok := globalCache[group]; ok {
		globallock.RUnlock()
		return m
	}
	globallock.RUnlock()
	m := &memoryCache{
		sctrl: make(chan struct{}),
		data:  make(map[string]*memoryCacheItem),
		group: group,
	}
	m.cleanup()
	globallock.Lock()
	globalCache[group] = m
	globallock.Unlock()
	return m
}

func MemoryCache(group string) Cache {
	return Memory(group)
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
	c.datalock.Lock()
	defer c.datalock.Unlock()
	for _, k := range keys {
		delete(c.data, k)
	}
	return nil
}

func (c *memoryCache) Close() error {
	c.sctrl <- struct{}{}
	close(c.sctrl)
	globallock.Lock()
	delete(globalCache, c.group)
	globallock.Unlock()
	return nil
}

func (c *memoryCache) ttl(ttl []time.Duration) time.Duration {
	if len(ttl) == 0 {
		return Permenent
	}
	return ttl[0]
}
