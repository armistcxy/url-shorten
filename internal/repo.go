package internal

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	bolt "go.etcd.io/bbolt"
)

/*
TODO: Implement 3 types of storage:
- Traditional: Postgres  => Done
- Embedded memory: SQLite
- K-V storage: Bolt

Redis ???
*/

type PostgresURLRepository struct {
	db *sqlx.DB
}

func NewPostgresURLRepository(dsn string) (*PostgresURLRepository, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	initTables(db)
	return &PostgresURLRepository{
		db: db,
	}, nil
}

func initTables(db *sqlx.DB) {
	createURLTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_urls_id ON urls (id);
	`
	_ = db.MustExec(createURLTableQuery)
}

func (pr *PostgresURLRepository) Create(ctx context.Context, url string) (*ShortURL, error) {
	id := randomString(6)
	insertURLQuery := `
		INSERT INTO urls (id, original_url) VALUES ($1, $2) RETURNING created_at;
	`
	short := &ShortURL{
		ID:     id,
		Origin: url,
	}
	if err := pr.db.GetContext(ctx, &short.CreatedAt, insertURLQuery, id, url); err != nil {
		return nil, err
	}
	return short, nil
}

func (pr *PostgresURLRepository) Get(ctx context.Context, id string) (string, error) {
	getURLQuery := `
		SELECT original_url FROM urls
		WHERE id=$1;
	`
	var origin string
	if err := pr.db.GetContext(ctx, &origin, getURLQuery, id); err != nil {
		return "", err
	}
	return origin, nil
}

type BoltURLRepository struct {
	db *bolt.DB
}

func NewBoltURLRepository(path string) (*BoltURLRepository, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &BoltURLRepository{
		db: db,
	}, nil
}

// TODO: using context in database operation

func (br *BoltURLRepository) Create(ctx context.Context, url string) (*ShortURL, error) {
	id := randomString(6)

	err := br.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("urls"))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(id), []byte(url))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ShortURL{ID: id, Origin: url}, nil
}

func (br *BoltURLRepository) Get(ctx context.Context, id string) (string, error) {
	var value []byte
	err := br.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("urls"))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		value = bucket.Get([]byte(id))
		return nil
	})

	if err != nil {
		return "", err
	}

	if value == nil {
		return "", fmt.Errorf("key %s not found in bucket", id)
	}

	return string(value), nil
}
