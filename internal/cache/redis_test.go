package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestViewCacheRedis(t *testing.T) {
	viewRedisCache := createViewRedisCache()

	testcases := []struct {
		testname string
		view     int
		expected int
		key      string
	}{
		{
			testname: "sequential",
			view:     3,
			expected: 3,
			key:      "url1",
		},
		{
			testname: "concurrent",
			view:     200,
			expected: 200,
			key:      "url2",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testname, func(t *testing.T) {
			if tc.testname == "sequential" {
				for range tc.view {
					err := viewRedisCache.Increase(context.Background(), tc.key)
					if err != nil {
						t.Error(err)
					}
				}
			} else {
				var wg sync.WaitGroup
				wg.Add(10)
				for range 10 {
					go func() {
						for range tc.view / 10 {
							err := viewRedisCache.Increase(context.Background(), tc.key)
							if err != nil {
								t.Error(err)
							}
						}
						wg.Done()
					}()
				}
				wg.Wait()
			}
			view, err := viewRedisCache.Get(context.Background(), tc.key)
			if err != nil {
				t.Errorf("Failed to retrieve view from key %s\n", err)
				return
			}
			assert.Equal(t, tc.expected, view)
		})
	}

	for _, tc := range testcases {
		cleanUp(viewRedisCache.client, tc.key)
	}
}

func createViewRedisCache() *ViewRedisCache {
	redisURLs := make([]string, 9)
	for i := 1; i <= 9; i++ {
		redisURLs[i-1] = fmt.Sprintf("redis://localhost:%d", i+6379)
	}
	return NewViewRedisCache(redisURLs)
}

func cleanUp(client *redis.ClusterClient, key string) {
	client.Del(context.Background(), key)
}
