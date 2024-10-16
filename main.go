package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/armistcxy/shorten/internal/handler"
	"github.com/armistcxy/shorten/internal/idgen"
	"github.com/armistcxy/shorten/internal/repository"
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

	addr := fmt.Sprintf("%s:%d", *host, *port)

	srv := http.Server{
		Addr:    addr,
		Handler: CORS(ApplyChain(http.DefaultServeMux, HTTPLoggingMiddleware)),
	}

	postgresDSN := os.Getenv("URL_DSN")
	postgresURLRepo, err := repository.NewPostgresURLRepository(postgresDSN)
	if err != nil {
		log.Fatal(err)
	}

	idgen := idgen.NewRandomIDGenerator()

	urlHandler := handler.NewURLHandler(postgresURLRepo, idgen)
	{
		createShortURLHandler := http.HandlerFunc(urlHandler.CreateShortURLHandle)
		http.Handle("POST /short", createShortURLHandler)

		getURLHandler := http.HandlerFunc(urlHandler.GetOriginURLHandle)
		http.Handle("GET /short/{id}", getURLHandler)
	}

	// Warning: This is just a work around to deal with my concurrent problem with load testing using my tool (https://github.com/armistcxy/go-load-testing)
	// My tool doesn't have dynamic query value feature yet !! Big update soon
	waCreateShortURLHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
			return
		}

		// create json payload
		payload := map[string]string{"origin": url}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, fmt.Sprintf("fail to create json payload, error: %s", err), http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/short", addr), bytes.NewBuffer(jsonPayload))
		if err != nil {
			http.Error(w, fmt.Sprintf("fail to create new request, error: %s", err), http.StatusInternalServerError)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("fail to forward request, error: %s", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, "fail to copy response body", http.StatusInternalServerError)
			return
		}
	})
	http.Handle("POST /create", waCreateShortURLHandler)
	// Gracefully shutdown
	done := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, os.Interrupt)
		<-quit

		waitTime := 1 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), waitTime)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Error when shutdown HTTP server", "error", err.Error())
		}

		close(done)
	}()

	log.Printf("Start listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error has ben occured in HTTP server ListenAndServe: %s", err.Error())
	}

	<-done
}
