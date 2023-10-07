package cache

import (
	"context"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	cache *cache.Cache
}

func RedisCache(url string) (Redis, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return Redis{}, err
	}
	rdb := redis.NewClient(opts)
	cache := cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return Redis{
		cache: cache,
	}, nil
}

func (r *Redis) Get(ctx context.Context, ip string, cachedResponse *CachedResponse) error {
	return r.cache.Get(ctx, ip, cachedResponse)
}

func (r *Redis) Set(ctx context.Context, ip string, response CachedResponse, cacheTtl int) error {
	return r.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   ip,
		Value: response,
		TTL:   time.Duration(cacheTtl * int(time.Second)),
	})
}
