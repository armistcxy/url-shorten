package repository

import (
	"context"

	"github.com/armistcxy/shorten/internal/domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

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
