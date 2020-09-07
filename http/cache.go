package http

import (
	"container/list"
	"hash/fnv"
	"net"
	"sync"
)

type Cache struct {
	capacity int
	mu       sync.RWMutex
	entries  map[uint64]Response
	keys     *list.List
}

func NewCache(capacity int) *Cache {
	if capacity < 0 {
		capacity = 0
	}
	keys := list.New()
	return &Cache{
		capacity: capacity,
		entries:  make(map[uint64]Response),
		keys:     keys,
	}
}

func key(ip net.IP) uint64 {
	h := fnv.New64a()
	h.Write(ip)
	return h.Sum64()
}

func (c *Cache) Set(ip net.IP, resp Response) {
	if c.capacity == 0 {
		return
	}
	k := key(ip)
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) == c.capacity {
		// At capacity. Remove the oldest entry
		oldest := c.keys.Front()
		delete(c.entries, oldest.Value.(uint64))
		c.keys.Remove(oldest)
	}
	c.entries[k] = resp
	c.keys.PushBack(k)
}

func (c *Cache) Get(ip net.IP) (Response, bool) {
	k := key(ip)
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.entries[k]
	return r, ok
}
