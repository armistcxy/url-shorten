package background

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

type AddLastUsedIDArgs struct {
	LastUsedID uint64
}

func (AddLastUsedIDArgs) Kind() string {
	return "add_last_used_id"
}

type AddLastUsedIDWorker struct {
	db *sqlx.DB
	river.WorkerDefaults[AddLastUsedIDArgs]
}

func NewAddLastUsedIDWorker(db *sqlx.DB) *AddLastUsedIDWorker {
	return &AddLastUsedIDWorker{
		db: db,
	}
}

func (aw *AddLastUsedIDWorker) Work(ctx context.Context, job *river.Job[AddLastUsedIDArgs]) error {
	return updateLastUsedID(aw.db, job.Args.LastUsedID)
}

// Mark this as a work for worker to make sure this operation must be done (retry if failed)
func updateLastUsedID(db *sqlx.DB, id uint64) error {
	query := "INSERT INTO ids(id) VALUES ($1) ON CONFLICT(id) DO NOTHING;"
	// don't know whether duplicate can happen, just to make sure everything works fine
	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	query = "SELECT COUNT(*) FROM ids WHERE id=$1"
	var count int
	if err = db.Get(&count, query, id); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("failed to insert ID %d", id)
	}
	slog.Info("Update last used id successfully", "id", id)
	return nil
}

func Migrate(ctx context.Context, dbPool *pgxpool.Pool) error {
	migrator, err := rivermigrate.New(riverpgxv5.New(dbPool), nil)
	if err != nil {
		return err
	}
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return err
	}
	return nil
}
