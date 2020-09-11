package http

import (
	"container/list"
	"fmt"
	"hash/fnv"
	"net"
	"sync"
)

type Cache struct {
	capacity  int
	mu        sync.RWMutex
	entries   map[uint64]*list.Element
	values    *list.List
	evictions uint64
}

type CacheStats struct {
	Capacity  int
	Size      int
	Evictions uint64
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
	minEvictions := len(c.entries) - c.capacity + 1
	if minEvictions > 0 { // At or above capacity. Shrink the cache
		evicted := 0
		for el := c.values.Front(); el != nil && evicted < minEvictions; {
			value := el.Value.(Response)
			delete(c.entries, key(value.IP))
			next := el.Next()
			c.values.Remove(el)
			el = next
			evicted++
		}
		c.evictions += uint64(evicted)
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

func (c *Cache) Resize(capacity int) error {
	if capacity < 0 {
		return fmt.Errorf("invalid capacity: %d\n", capacity)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.capacity = capacity
	c.evictions = 0
	return nil
}

func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		Size:      len(c.entries),
		Capacity:  c.capacity,
		Evictions: c.evictions,
	}
}
