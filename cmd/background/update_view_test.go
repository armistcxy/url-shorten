package main

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestBatchUpdateView(t *testing.T) {
	db := sqlx.MustConnect("postgres", os.Getenv("URL_DSN"))
	snapshotItems := []SnapshotItem{{Key: "b", Value: 56}, {Key: "a", Value: 12}}
	err := batchUpdateView(db, snapshotItems)
	if err != nil {
		t.Error(err)
	}

	expectedValues := map[string]int{}
	for _, item := range snapshotItems {
		expectedValues[item.Key] = item.Value
	}

	for key, expected := range expectedValues {
		view, err := getView(db, key)
		if err != nil {
			t.Error(err)
			continue
		}
		assert.Equal(t, expected, view)
	}
}

func getView(db *sqlx.DB, id string) (int, error) {
	var (
		getViewQuery = `
			SELECT count
			FROM urls
			WHERE id=$1
		`
		count int
	)

	if err := db.GetContext(context.Background(), &count, getViewQuery, id); err != nil {
		return -1, nil
	}
	return count, nil
}
