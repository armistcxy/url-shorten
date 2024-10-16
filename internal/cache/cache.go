package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, id string) (string, error)
	Set(ctx context.Context, id string, url string) error
	SetWithTTL(ctx context.Context, id string, url string, ttl time.Duration) error
}
