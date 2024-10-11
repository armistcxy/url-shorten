package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

func init() {
	dsn := os.Getenv("URL_DSN")
	db := sqlx.MustConnect("postgres", dsn)
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

type URLFixture struct {
	ID     string `db:"id"`
	Origin string `db:"original_url"`
}

func TestPostgresRepoCreateShort(t *testing.T) {
	dsn := os.Getenv("URL_DSN")
	db := sqlx.MustConnect("postgres", dsn)
	fixtures := []URLFixture{
		{Origin: "https://www.example.com"},
		{Origin: "https://www.google.com"},
		{Origin: "https://github.com"},
		{Origin: "https://stackoverflow.com"},
		{Origin: "https://www.youtube.com"},
	}

	// clean
	defer func() {
		if err := postgresRepoCleanHelper(db, fixtures); err != nil {
			t.Fatal(err)
		}
	}()

	repo, err := NewPostgresURLRepository(dsn)
	if err != nil {
		t.Fatal(err)
	}

	type TestCase struct {
		testName string
		fixture  URLFixture
	}

	testcases := make([]TestCase, len(fixtures))

	// prepare testcases
	for i, ft := range fixtures {
		testcases[i] = TestCase{
			testName: fmt.Sprintf("Test %d", i+1),
			fixture:  ft,
		}
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			short, err := repo.Create(context.Background(), tc.fixture.Origin)
			if err != nil {
				t.Errorf("fail to retrieve create short url, error: %s", err)
				return
			}
			assert.Equal(t, tc.fixture.Origin, short.Origin)
		})
	}
}

func TestPostgresRepoGetOrigin(t *testing.T) {
	dsn := os.Getenv("URL_DSN")
	db := sqlx.MustConnect("postgres", dsn)
	fixtures := []URLFixture{
		{Origin: "https://www.example.com"},
		{Origin: "https://www.google.com"},
		{Origin: "https://github.com"},
		{Origin: "https://stackoverflow.com"},
		{Origin: "https://www.youtube.com"},
	}
	for i := range fixtures {
		fixtures[i].ID = randomString(6)
	}

	// clean
	defer func() {
		if err := postgresRepoCleanHelper(db, fixtures); err != nil {
			t.Fatal(err)
		}
	}()

	// init
	if err := postgresRepoInitHelper(db, fixtures); err != nil {
		t.Fatal(err)
	}

	repo, err := NewPostgresURLRepository(dsn)
	if err != nil {
		t.Fatal(err)
	}

	type TestCase struct {
		testName string
		fixture  URLFixture
	}

	testcases := make([]TestCase, len(fixtures))

	// prepare testcases
	for i, ft := range fixtures {
		testcases[i] = TestCase{
			testName: fmt.Sprintf("Test %d", i+1),
			fixture:  ft,
		}
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			origin, err := repo.Get(context.Background(), tc.fixture.ID)
			if err != nil {
				t.Errorf("fail to retrieve origin url, error: %s", err)
				return
			}
			assert.Equal(t, tc.fixture.Origin, origin)
		})
	}
}

func BenchmarkPostgresRepositoryMassiveCreate(b *testing.B) {
	b.StopTimer()
	var short *ShortURL
	var err error
	dsn := os.Getenv("URL_DSN")
	repo, err := NewPostgresURLRepository(dsn)
	if err != nil {
		slog.Error("failed when prepare repository for benchmark massive create (postgresql)", "error", err.Error())
		b.FailNow()
	}
	ids := make([]string, 0)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		url := faker.URL()
		b.StartTimer()

		short, err = repo.Create(ctx, url)
		if err != nil {
			slog.Error("failed to create short url", "error", err.Error())
		}

		b.StopTimer()
		ids = append(ids, short.ID)
		b.StartTimer()
	}

	b.StopTimer()
	b.Cleanup(func() {
		query := `
			DELETE FROM urls
			WHERE id=$1
		`
		for id := range ids {
			_, err := repo.db.Exec(query, id)
			if err != nil {
				slog.Error("failed when cleaning up", "error", err.Error())
			}
		}
	})
}

func postgresRepoInitHelper(db *sqlx.DB, fixtures []URLFixture) error {
	// bulk insert with db.NamedExec, (acutally not bulk insert feature but it save number of connections, 1 query add all)
	_, err := db.NamedExec(`INSERT INTO urls (id, original_url) VALUES (:id, :original_url)`, fixtures)
	return err
}

func postgresRepoCleanHelper(db *sqlx.DB, fixtures []URLFixture) error {
	deleteQuery := `
		DELETE FROM urls 
		WHERE original_url=$1
	`

	for _, fixture := range fixtures {
		_, err := db.Exec(deleteQuery, fixture.Origin)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestBoltRepositoryCreateShort(t *testing.T) {}
func TestBoltRepositoryGetOrigin(t *testing.T)   {}
func BenchmarkBoltRepository(b *testing.B)       {}

func TestSQLiteRepositoryCreateShort(t *testing.T) {}
func TestSQLiteRepositoryGetOrigin(t *testing.T)   {}
func BenchmarkSQLiteRepository(b *testing.B)       {}

// honour mention: Redis ??
