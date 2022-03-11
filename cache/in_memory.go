// Because who doesn't love to write their own cache implementation huehuehue
package cache

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

var _ CacheProvider = &InMemoryCache{}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

type InMemoryCache struct {
	logger           *zap.Logger
	cacheMap         map[string]cacheEntry
	spawnJanitorOnce sync.Once
	mu               sync.RWMutex
	defaultCacheTTL  time.Duration
}

func (c *InMemoryCache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	val, ok := c.cacheMap[k]
	c.mu.RUnlock()
	return val.data, ok
}

func (c *InMemoryCache) Delete(k string) {
	c.mu.Lock()
	delete(c.cacheMap, k)
	c.mu.Unlock()
}

func (c *InMemoryCache) Set(k string, data interface{}) {
	c.mu.Lock()
	c.cacheMap[k] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.defaultCacheTTL),
	}
	c.mu.Unlock()
	c.SpawnJanitorGoroutine()
}

func (c *InMemoryCache) SpawnJanitorGoroutine() {
	c.spawnJanitorOnce.Do(func() {
		go func() {
			for {
				c.mu.Lock()
				janitoredData := make([]interface{}, 0)
				for k, v := range c.cacheMap {
					if v.expiresAt.Before(time.Now()) {
						delete(c.cacheMap, k)
						janitoredData = append(janitoredData, v)
					}
				}
				c.mu.Unlock()
				if c.logger != nil && len(janitoredData) > 0 {
					c.logger.Debug(
						"Removed expired cached values.",
						zap.Any("Values", janitoredData),
					)
				}
				time.Sleep(1 * time.Minute)
			}
		}()
	})
}

// NewInMemoryCache creates a new cache. pass in a nil logger if you don't want to log anything.
func NewInMemoryCache(logger *zap.Logger, label string) CacheProvider {
	c := &InMemoryCache{}
	if logger != nil {
		c.logger = logger.Named("cache").Named(label)
	}
	return c
}
