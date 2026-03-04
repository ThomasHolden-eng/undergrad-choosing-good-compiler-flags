package util

import "sync"

// Cache is a basic map designed for concurrent use.
type Cache[K, V comparable] struct {
	cache         map[K]V    // The underlying map for the cache struct
	queue         []K        // The queue of keys
	queueLastPtr  int        // The pointer to the last element in the queue
	queueFirstPtr int        // The pointer to the first element in the queue
	lock          sync.Mutex // The lock for concurrent access to the cache
}

// Initialises the map components.
func (c *Cache[K, V]) Init(size int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cache == nil {
		c.cache = make(map[K]V)
		c.queue = make([]K, size)
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
	_, inMap := c.cache[key]
	c.cache[key] = value

	if !inMap {
		c.enqueue(key)
	}
}

func (c *Cache[K, V]) enqueue(key K) {
	if len(c.cache) > cap(c.queue) {
		keyToRemove := c.queue[c.queueFirstPtr]
		delete(c.cache, keyToRemove)
		c.queueFirstPtr = (c.queueFirstPtr + 1) % cap(c.queue)
	}
	c.queue[c.queueLastPtr] = key
	c.queueLastPtr = (c.queueLastPtr + 1) % cap(c.queue)
}
