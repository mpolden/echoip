package http

import (
	"container/list"
	"fmt"
	"hash/fnv"
	"net"
	"sync"
)

type Cache struct {
	capacity int
	mu       sync.RWMutex
	entries  map[uint64]*list.Element
	values   *list.List
}

type CacheStats struct {
	Capacity int
	Size     int
}

func NewCache(capacity int) *Cache {
	if capacity < 0 {
		capacity = 0
	}
	return &Cache{
		capacity: capacity,
		entries:  make(map[uint64]*list.Element),
		values:   list.New(),
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
		oldest := c.values.Front()
		oldestValue := oldest.Value.(Response)
		oldestKey := key(oldestValue.IP)
		delete(c.entries, oldestKey)
		c.values.Remove(oldest)
	}
	current, ok := c.entries[k]
	if ok {
		c.values.Remove(current)
	}
	c.entries[k] = c.values.PushBack(resp)
}

func (c *Cache) Get(ip net.IP) (Response, bool) {
	k := key(ip)
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.entries[k]
	if !ok {
		return Response{}, false
	}
	return r.Value.(Response), true
}

func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		Size:     len(c.entries),
		Capacity: c.capacity,
	}
}
