package cache

import (
	"context"
)

type Null struct{}

func (nc *Null) Get(ctx context.Context, ip string, cachedResponse *CachedResponse) error {
	return nil
}

func (nc *Null) Set(ctx context.Context, ip string, response CachedResponse, cacheTtl int) error {
	return nil
}
