package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/armistcxy/shorten/internal/domain"
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

func (pr *PostgresURLRepository) Create(ctx context.Context, id string, url string) (*domain.ShortURL, error) {
	insertURLQuery := `
		INSERT INTO urls (id, original_url) VALUES ($1, $2) RETURNING created_at;
	`
	short := &domain.ShortURL{
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

// K-V storage approach start from here
// Current approach: Divide into partitions, each partitions will responsible for a
// specific range, each partition will be stored inside one bucket
type BoltURLRepository struct {
	db *bolt.DB
}

const (
	PARTITION_SIZE = 1 << 25
)

var bboltPath = os.Getenv("KV_STORAGE_PATH")

func NewBoltURLRepository(path string) (*BoltURLRepository, error) {
	db, err := bolt.Open(bboltPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &BoltURLRepository{
		db: db,
	}, nil
}

// bbolt can't perform concurrent write and that's bad => I will consider switch to another storage
// bbolt will turn concurrent write operations into serialization
func (br *BoltURLRepository) Create(ctx context.Context, id string, url string) (*domain.ShortURL, error) {
	decodeID := domain.DecodeID(id)
	bucketName := fmt.Sprintf("short-%d", decodeID/PARTITION_SIZE)
	err := br.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return newErrBucketNotExist(bucketName)
		}
		return b.Put([]byte(id), []byte(url))
	})
	if err != nil {
		return nil, err
	}
	shortenURL := &domain.ShortURL{
		ID:        id,
		Origin:    url,
		CreatedAt: time.Now(),
	}
	return shortenURL, nil
}

func (br *BoltURLRepository) Get(ctx context.Context, id string) (string, error) {
	decodeID := domain.DecodeID(id)
	bucketName := fmt.Sprintf("short-%d", decodeID/PARTITION_SIZE)

	var origin string

	err := br.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return newErrBucketNotExist(bucketName)
		}
		result := b.Get([]byte(id))
		if result == nil {
			return newErrIDNotExist(id)
		}
		origin = string(result)
		return nil
	})

	if err != nil {
		return "", err
	}

	return origin, err
}

type ErrIDNotExist struct {
	id string
}

func newErrIDNotExist(id string) ErrIDNotExist {
	return ErrIDNotExist{id}
}

func (e ErrIDNotExist) Error() string {
	return fmt.Sprintf("there's no entry with id: %s", e.id)
}

type ErrBucketNotExist struct {
	name string
}

func newErrBucketNotExist(name string) ErrBucketNotExist {
	return ErrBucketNotExist{name}
}

func (e ErrBucketNotExist) Error() string {
	return fmt.Sprintf("there's no bucket with name: %s", e.name)
}
