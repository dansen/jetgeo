package geo

import (
	"sync"
	"time"
)

type cacheEntry struct {
	regions  []*RegionCache
	loadedAt time.Time
}

type LoadingCache struct {
	mu     sync.RWMutex
	data   map[string]*cacheEntry
	ttl    time.Duration
	loader func(key string) ([]*RegionCache, error)
}

func NewLoadingCache(ttl time.Duration, loader func(string) ([]*RegionCache, error)) *LoadingCache {
	return &LoadingCache{
		data:   make(map[string]*cacheEntry),
		ttl:    ttl,
		loader: loader,
	}
}

func (c *LoadingCache) Get(key string) ([]*RegionCache, error) {
	c.mu.RLock()
	entry, ok := c.data[key]
	c.mu.RUnlock()
	if ok && time.Since(entry.loadedAt) < c.ttl {
		return entry.regions, nil
	}
	// load or reload
	regions, err := c.loader(key)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.data[key] = &cacheEntry{regions: regions, loadedAt: time.Now()}
	c.mu.Unlock()
	return regions, nil
}
