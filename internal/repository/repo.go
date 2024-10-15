package repository

import (
	"context"

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

func (pr *PostgresURLRepository) Create(ctx context.Context, url string) (*domain.ShortURL, error) {
	id := domain.RandomString(6)
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

type BoltURLRepository struct {
	db         *bolt.DB
	partitions []Partition // all id inside partition range stored in one bucket
}

const (
	PARTITION_SIZE int   = 1 << 25 // random number, don't mind
	BASE           int64 = 62
)

// A partition considered archieved if all the id inside its range have been used
// (i.e., used == PARTITION_SIZE)
type Partition struct {
	start int // Start id in partition
	used  int // Number of id have already been used
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

func (br *BoltURLRepository) Create(ctx context.Context, url string) (*domain.ShortURL, error) {
	return nil, nil
}

func (br *BoltURLRepository) Get(ctx context.Context, id string) (string, error) {
	return "", nil
}
