package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type SnapshotItem struct {
	Key   string
	Value int
}

func TakeSnapshot(rdb *redis.ClusterClient) ([]SnapshotItem, error) {
	var items []SnapshotItem

	err := rdb.ForEachMaster(context.Background(), func(ctx context.Context, master *redis.Client) error {
		var cursor uint64

		for {
			keys, nextCursor, err := master.Scan(ctx, cursor, "view:*", 1000).Result()
			if err != nil {
				return fmt.Errorf("error scanning keys: %w", err)
			}

			if len(keys) > 0 {
				values, err := master.MGet(ctx, keys...).Result()
				if err != nil {
					return fmt.Errorf("error fetching values: %w", err)
				}

				for i, key := range keys {
					if values[i] == nil {
						continue
					}
					val, ok := values[i].(string)
					if !ok {
						continue
					}

					intVal, err := strconv.Atoi(val)
					if err != nil {
						continue
					}

					items = append(items, SnapshotItem{
						Key:   strings.TrimPrefix(key, "view:"),
						Value: intVal,
					})
				}
			}

			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}
