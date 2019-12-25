package http

import (
	"fmt"
	"net"
	"testing"
)

func TestCacheCapacity(t *testing.T) {
	var tests = []struct {
		addCount, capacity, size int
	}{
		{1, 0, 0},
		{1, 2, 1},
		{2, 2, 2},
		{3, 2, 2},
	}
	for i, tt := range tests {
		c := NewCache(tt.capacity)
		var responses []*Response
		for i := 0; i < tt.addCount; i++ {
			ip := net.ParseIP(fmt.Sprintf("192.0.2.%d", i))
			r := &Response{IP: ip}
			responses = append(responses, r)
			c.Set(ip, r)
		}
		if got := len(c.entries); got != tt.size {
			t.Errorf("#%d: len(entries) = %d, want %d", i, got, tt.size)
		}
		if tt.capacity > 0 && tt.addCount > tt.capacity && tt.capacity == tt.size {
			lastAdded := responses[tt.addCount-1]
			if _, ok := c.Get(lastAdded.IP); !ok {
				t.Errorf("#%d: Get(%s) = (_, %t), want (_, %t)", i, lastAdded.IP.String(), ok, !ok)
			}
			firstAdded := responses[0]
			if _, ok := c.Get(firstAdded.IP); ok {
				t.Errorf("#%d: Get(%s) = (_, %t), want (_, %t)", i, firstAdded.IP.String(), ok, !ok)
			}
		}
	}
}
