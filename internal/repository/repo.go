package repository

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/armistcxy/shorten/internal/domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresURLRepository struct {
	db *sqlx.DB
}

func NewPostgresURLRepository(db *sqlx.DB) (*PostgresURLRepository, error) {
	initTables(db)
	return &PostgresURLRepository{
		db: db,
	}, nil
}

// Not encourage to do this
// Note that this is redundant work just to make sure that db schema is up to date
func initTables(db *sqlx.DB) {
	createURLTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			fraud BOOLEAN DEFAULT false,
			count INTEGER DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_urls_id ON urls (id);

		CREATE TABLE IF NOT EXISTS ids (
			id BIGINT PRIMARY KEY
		);
	`
	_ = db.MustExec(createURLTableQuery)
}

var (
	insertURLQuery = `
		INSERT INTO urls (id, original_url) VALUES ($1, $2) RETURNING created_at;
	`
)

func (pr *PostgresURLRepository) Create(ctx context.Context, id string, url string) (*domain.ShortURL, error) {

	// Consideration: Removing `created_at` field
	short := &domain.ShortURL{
		ID:     id,
		Origin: url,
	}
	if err := pr.db.GetContext(ctx, &short.CreatedAt, insertURLQuery, id, url); err != nil {
		return nil, err
	}
	return short, nil
}

var (
	getURLQuery = `
		SELECT original_url FROM urls
		WHERE id=$1;
	`
)

func (pr *PostgresURLRepository) Get(ctx context.Context, id string) (string, error) {
	var origin string
	if err := pr.db.GetContext(ctx, &origin, getURLQuery, id); err != nil {
		return "", err
	}
	return origin, nil
}

var (
	retrieveFraudQuery = `
		SELECT fraud FROM urls
		WHERE id=$1
	`
)

func (pr *PostgresURLRepository) RetrieveFraud(ctx context.Context, id string) (bool, error) {
	var fraud bool
	if err := pr.db.GetContext(ctx, &fraud, retrieveFraudQuery, id); err != nil {
		return false, err
	}
	return fraud, nil
}

var (
	getViewQuery = `
		SELECT count
		FROM urls
		WHERE id=$1
	`
)

func (pr *PostgresURLRepository) GetView(ctx context.Context, id string) (int, error) {
	var view int
	if err := pr.db.GetContext(ctx, &view, getViewQuery, id); err != nil {
		return 0, err
	}
	return view, nil
}

func (pr *PostgresURLRepository) BatchCreate(ctx context.Context, inputs []domain.CreateInput) error {
	byteBuffer := bytes.NewBufferString(`INSERT INTO urls (id, original_url) VALUES `)
	tuples := make([]string, len(inputs))
	for i := range tuples {
		tuples[i] = prepareEach(inputs[i])
	}
	byteBuffer.WriteString(strings.Join(tuples, ","))

	query := byteBuffer.String()
	if _, err := pr.db.Exec(query); err != nil {
		return err
	}
	return nil
}

func prepareEach(input domain.CreateInput) string {
	return fmt.Sprintf("('%s', '%s')", input.ID, input.URL)
}
