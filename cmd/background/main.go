package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/armistcxy/shorten/internal/background"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func main() {
	waitForOtherServicesDuration := 10 * time.Second
	time.Sleep(waitForOtherServicesDuration)

	riverDSN := os.Getenv("RIVER_DSN")
	dbPool, err := pgxpool.New(context.Background(), riverDSN)
	if err != nil {
		log.Fatal(err)
	}

	if err := background.Migrate(context.Background(), dbPool); err != nil {
		log.Fatal(err)
	}

	urlDSN := os.Getenv("URL_DSN")
	db := sqlx.MustConnect("postgres", urlDSN)

	redisURLs := make([]string, 9)
	for i := 1; i <= 9; i++ {
		redisURLs[i-1] = fmt.Sprintf("redis://redis_%d:6379", i)
		// redisURLs[i-1] = fmt.Sprintf("redis://localhost:%d", i+6379)
	}
	parsedURLs := make([]string, len(redisURLs))
	for i := range parsedURLs {
		if opt, err := redis.ParseURL(redisURLs[i]); err != nil {
			panic(err)
		} else {
			parsedURLs[i] = opt.Addr
		}
	}
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: parsedURLs,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to ping Redis cluster: %v\n", err)
	} else {
		log.Println("Successfully connected to Redis cluster")
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, background.NewAddLastUsedIDWorker(db))

	batchCreateWorker := background.NewBatchCreateWorker(db)
	river.AddWorker(workers, batchCreateWorker)

	updateViewWorker := NewUpdateViewWorker(5*time.Second, db, client)
	go updateViewWorker.Work()

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})

	if err != nil {
		log.Fatal(err)
	}

	if err = riverClient.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Println("Background worker has started working")
	log.Fatal(http.ListenAndServe(":8010", nil))
}
