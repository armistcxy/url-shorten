package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/armistcxy/shorten/internal/background"
	"github.com/armistcxy/shorten/internal/cache"
	"github.com/armistcxy/shorten/internal/domain"
	"github.com/armistcxy/shorten/internal/msq"
	"github.com/armistcxy/shorten/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// This will deal with 2 end points
// /short/:id GET => Return original url
// /create?url= POST => Return short url
type URLHandler struct {
	urlRepo     domain.URLRepository
	idGen       domain.IDGenerator
	cache       cache.Cache
	pub         *msq.URLPublisher
	riverClient *river.Client[pgx.Tx]
}

func NewURLHandler(urlRepo domain.URLRepository, idGen domain.IDGenerator, cache cache.Cache, pub *msq.URLPublisher, riverClient *river.Client[pgx.Tx]) *URLHandler {
	return &URLHandler{
		urlRepo:     urlRepo,
		idGen:       idGen,
		cache:       cache,
		pub:         pub,
		riverClient: riverClient,
	}
}

// GetOriginURLHandle handles the GET request to retrieve the original URL for a given short URL ID.
// It extracts the ID from the request path, looks up the original URL in the URLRepository,
// and encodes the original URL as a JSON response.
func (uh *URLHandler) GetOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]

	cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	originURL, err := uh.cache.Get(cacheCtx, id)
	if err != nil {
		slog.Error("failed when trying to retrieve entry from cache", "error", err.Error())
	} else if originURL != "" {
		util.EncodeJSON(w, map[string]string{"origin": originURL})
		return
	}

	originURL, err = uh.urlRepo.Get(context.Background(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("fail to retrive origin url, error: %s", err), http.StatusInternalServerError)
		return
	}
	util.EncodeJSON(w, map[string]string{"origin": originURL})
}

// CreateShortURLHandle handles the POST request to create a new short URL.
// It extracts the original URL from the request, creates a new short URL using the URLRepository,
// and encodes the short URL as a JSON response.
func (uh *URLHandler) CreateShortURLHandle(w http.ResponseWriter, r *http.Request) {
	form := CreateShortForm{}
	if err := util.DecodeJSON(r, &form); err != nil {
		slog.Error("fail when decoding json body", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uh.idGen.GenerateID()

	short, err := uh.urlRepo.Create(context.Background(), id, form.Origin)
	if err != nil {
		slog.Error("failed to insert url to database", "error", err.Error())
		http.Error(w, fmt.Sprintf("failed when creating short url, error: %s", err), http.StatusInternalServerError)
		return
	}

	// Add k-v pair (id:origin_url) to cache for 5 minutes
	cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := uh.cache.SetWithTTL(cacheCtx, id, form.Origin, 5*time.Minute); err != nil {
		slog.Error("failed to set k-v to cache", "id", id, "origin", form.Origin, "error", err.Error())
	}

	go func() {
		if err := uh.pub.EnqueueURL(context.Background(), form.Origin, id); err != nil {
			slog.Error("failed to enequeue url", "url", form.Origin, "url_id", id, "error", err.Error())
		}
		if _, err := uh.riverClient.Insert(context.Background(), background.IncreaseCountArgs{
			ID: id,
		}, nil); err != nil {
			slog.Error("failed to enqueue increase view jobs", "url_id", id, "error", err.Error())
		}
	}()

	util.EncodeJSON(w, short)
}

type CreateShortForm struct {
	Origin string `json:"origin"`
}

func (uh *URLHandler) RetrieveFraudURLHandle(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]

	fraud, err := uh.urlRepo.RetrieveFraud(context.Background(), id)
	if err != nil {
		slog.Error("fail to retrieve fraud from database", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.EncodeJSON(w, map[string]interface{}{"fraud": fraud})
}
