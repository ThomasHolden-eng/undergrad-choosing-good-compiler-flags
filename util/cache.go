package util

import "sync"

// Cache is a basic map designed for concurrent use.
type Cache[K, V comparable] struct {
	cache map[K]V    // The underlying map for the cache struct
	lock  sync.Mutex // The lock for concurrent access to the cache
}

// Initialises the map components.
func (c *Cache[K, V]) Init() {
	if c.cache == nil {
		c.cache = make(map[K]V)
	}
}

// Get accesses the map by key to obtain a value.
//
// Returns the value and whether it was found.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	value, ok := c.cache[key]
	return value, ok
}

// Set adds a key value pair to the map.
func (c *Cache[K, V]) Set(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache[key] = value
}
