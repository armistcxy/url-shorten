package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/armistcxy/shorten/internal/background"
	"github.com/armistcxy/shorten/internal/cache"
	"github.com/armistcxy/shorten/internal/handler"
	"github.com/armistcxy/shorten/internal/idgen"
	"github.com/armistcxy/shorten/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// _ = make([]byte, 10<<30)
	host := flag.String("host", "", "Host of HTTP server")
	port := flag.Int("port", 8080, "Port that HTTP server listen to")

	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Llongfile)

	store, err := memorystore.New(&memorystore.Config{
		Tokens:   10,
		Interval: time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Rate limit based on IP address (IP is taken from "X-Forwareded-For" header)
	// https://pkg.go.dev/github.com/sethvargo/go-limiter/httplimit#IPKeyFunc
	rateLimitMiddleware, err := httplimit.NewMiddleware(store, httplimit.IPKeyFunc("X-Forwarded-For"))
	if err != nil {
		log.Fatal(err)
	}

	// consider using RateLimit in Traefik: https://doc.traefik.io/traefik/middlewares/http/ratelimit/

	var (
		addr = fmt.Sprintf("%s:%d", *host, *port)
		srv  = http.Server{
			Addr:    addr,
			Handler: rateLimitMiddleware.Handle((ApplyChain(http.DefaultServeMux, HTTPLoggingMiddleware))),
		}
	)

	dbPool, err := pgxpool.New(context.Background(), os.Getenv("RIVER_DSN"))
	if err != nil {
		slog.Error("failed to create database pool", "error", err.Error())
	}

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
	if err != nil {
		panic(err)
	}

	db := sqlx.MustConnect("postgres", os.Getenv("URL_DSN"))

	postgresURLRepo, err := repository.NewPostgresURLRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	redisURL := os.Getenv("REDIS_URL")
	ca := cache.NewRedisCache(redisURL)

	idgen := idgen.NewSeqIDGenerator(db, 0, 12, riverClient)

	urlHandler := handler.NewURLHandler(postgresURLRepo, idgen, nil, ca)
	{
		createShortURLHandler := http.HandlerFunc(urlHandler.CreateShortURLHandle)
		http.Handle("POST /short", createShortURLHandler)

		getURLHandler := http.HandlerFunc(urlHandler.GetOriginURLHandle)
		http.Handle("GET /short/{id}", getURLHandler)
	}

	// Gracefully shutdown
	done := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		waitTime := 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), waitTime)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Error when shutdown HTTP server", "error", err.Error())
		}

		// After handling all the remain requests: update maximum ID for each range
		updateIDs := idgen.RetriveLastUsedIds()
		for _, id := range updateIDs {
			if id != 0 {
				_, err := riverClient.Insert(context.Background(),
					background.AddLastUsedIDArgs{LastUsedID: id}, nil)
				if err != nil {
					slog.Error("failed to enqueue 'AddLastUsedID' job", "error", err.Error())
				}
			}
		}
		close(done)
	}()

	log.Printf("Start listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error has ben occured in HTTP server ListenAndServe: %s", err.Error())
	}

	<-done
}
