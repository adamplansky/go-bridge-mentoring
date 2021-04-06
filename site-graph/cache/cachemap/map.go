package cachemap

import (
	"sync"
)

// cmap implements cache interface and use golang map as underlying
// data struct for storing objects

type Cache struct {
	cache map[string]interface{}
	// https://stackoverflow.com/questions/53427824/what-is-the-difference-between-rlock-and-lock-in-golang
	mu sync.RWMutex
}

func New() *Cache {
	return &Cache{
		cache: make(map[string]interface{}),
		mu:    sync.RWMutex{},
	}
}

func (c *Cache) Add(key string, value interface{}) {
	c.mu.Lock()
	c.cache[key] = value
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (value interface{}, ok bool) {
	c.mu.RLock()
	value, ok = c.cache[key]
	c.mu.RUnlock()
	return value, ok
}
