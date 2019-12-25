package http

import (
	"hash/fnv"
	"net"
	"sync"
)

type Cache struct {
	capacity int
	mu       sync.RWMutex
	entries  map[uint64]*Response
	keys     []uint64
}

func NewCache(capacity int) *Cache {
	if capacity < 0 {
		capacity = 0
	}
	return &Cache{
		capacity: capacity,
		entries:  make(map[uint64]*Response),
		keys:     make([]uint64, 0, capacity),
	}
}

func key(ip net.IP) uint64 {
	h := fnv.New64a()
	h.Write(ip)
	return h.Sum64()
}

func (c *Cache) Set(ip net.IP, resp *Response) {
	if c.capacity == 0 {
		return
	}
	k := key(ip)
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) == c.capacity && c.capacity > 0 {
		delete(c.entries, c.keys[0])
		c.keys = c.keys[1:]
	}
	c.entries[k] = resp
	c.keys = append(c.keys, k)
}

func (c *Cache) Get(ip net.IP) (*Response, bool) {
	k := key(ip)
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.entries[k]
	return r, ok
}
