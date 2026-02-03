package util

import "sync"

// Cache is a basic map designed for concurrent use.
type Cache[K, V comparable] struct {
	cache map[K]V
	lock  sync.Mutex
}

// Initialises the map components.
func (c *Cache[K, V]) Init() {
	if c.cache == nil {
		c.cache = make(map[K]V)
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	value, ok := c.cache[key]
	return value, ok
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache[key] = value
}
