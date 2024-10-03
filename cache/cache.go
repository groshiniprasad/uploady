package cache

import (
	"sync"
	"time"
)

// Cache struct holds the cached data and a mutex for thread safety
type Cache struct {
	data map[string]cacheValue
	lock sync.Mutex
}

// cacheValue holds the actual value and its expiration time
type cacheValue struct {
	value      interface{}
	expiration time.Time
}

// NewCache creates a new Cache instance and starts a background cleanup for expired items
func NewCache() *Cache {
	cache := &Cache{
		data: make(map[string]cacheValue),
	}
	go cache.cleanupExpired() // Start background cleanup for expired items
	return cache
}

// Set adds a new item to the cache with the specified expiration duration
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	expirationTime := time.Now().Add(expiration)
	c.data[key] = cacheValue{
		value:      value,
		expiration: expirationTime,
	}
}

// Get retrieves an item from the cache if it exists and has not expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	value, ok := c.data[key]
	if !ok || time.Now().After(value.expiration) {
		delete(c.data, key) // Remove expired item
		return nil, false
	}
	return value.value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.data, key)
}

// cleanupExpired runs in the background and periodically removes expired items
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute) // Cleanup every minute
	for range ticker.C {
		c.lock.Lock()
		for key, value := range c.data {
			if time.Now().After(value.expiration) {
				delete(c.data, key) // Remove expired item
			}
		}
		c.lock.Unlock()
	}
}

func (c *Cache) Size() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return len(c.data)
}
