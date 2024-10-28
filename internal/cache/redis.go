package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) *RedisCache {
	// There is an error when using directly redis URL as address
	// Use ParseURL instead as suggest from the issue below
	// https://github.com/redis/go-redis/issues/864
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: opt.Addr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = client.Ping(ctx).Err(); err != nil {
		panic(err)
	}
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
