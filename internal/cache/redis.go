package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{client: client}
}

func (rc *RedisCache) Get(ctx context.Context, id string) (string, error) {
	val, err := rc.client.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func (rc *RedisCache) Set(ctx context.Context, id string, url string) error {
	return rc.client.Set(ctx, id, url, 0).Err()
}

func (rc *RedisCache) SetWithTTL(ctx context.Context, id string, url string, ttl time.Duration) error {
	return rc.client.Set(ctx, id, url, ttl).Err()
}
