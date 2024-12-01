package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type UpdateViewWorker struct {
	period time.Duration
	db     *sqlx.DB
	rdb    *redis.ClusterClient
}

func NewUpdateViewWorker(period time.Duration, db *sqlx.DB, rdb *redis.ClusterClient) *UpdateViewWorker {
	return &UpdateViewWorker{
		period: period,
		db:     db,
		rdb:    rdb,
	}
}

func (uw *UpdateViewWorker) Work() {
	ticker := time.NewTicker(uw.period)
	defer ticker.Stop()

	for range ticker.C {
		snapshotItems, err := TakeSnapshot(uw.rdb)
		if err != nil {
			slog.Error("Failed to take snapshot", "error", err.Error())
			continue
		}
		batchUpdateView(uw.db, snapshotItems)

	}
}

var (
	batchUpdateQuery = `
		UPDATE urls AS u
		SET count = c.new_count::integer
		FROM (VALUES %s) AS c(id, new_count)
		WHERE u.id = c.id;
	`
)

func batchUpdateView(db *sqlx.DB, snapshotItemms []SnapshotItem) error {
	if len(snapshotItemms) == 0 {
		return nil
	}

	values := []string{}
	params := []interface{}{}

	for i, item := range snapshotItemms {
		values = append(values, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		params = append(params, item.Key, item.Value)
		i += 2
	}

	if len(params) == 0 {
		return nil
	}

	valuesClause := strings.Join(values, ",")
	query := fmt.Sprintf(batchUpdateQuery, valuesClause)

	_, err := db.ExecContext(context.Background(), query, params...)
	if err != nil {
		return err
	}

	return nil
}
