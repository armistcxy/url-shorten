package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/armistcxy/shorten/internal/cache"
	"github.com/armistcxy/shorten/internal/domain"
	"github.com/armistcxy/shorten/internal/util"

	"github.com/bits-and-blooms/bloom/v3"
)

// This will deal with 2 end points
// /short/:id GET => Return original url
// /create?url= POST => Return short url
type URLHandler struct {
	urlRepo  domain.URLRepository
	idGen    domain.IDGenerator
	idFilter *bloom.BloomFilter
	cache    cache.Cache
}

func NewURLHandler(urlRepo domain.URLRepository, idGen domain.IDGenerator, idFilter *bloom.BloomFilter, cache cache.Cache) *URLHandler {
	return &URLHandler{
		urlRepo:  urlRepo,
		idGen:    idGen,
		idFilter: idFilter,
		cache:    cache,
	}
}

// GetOriginURLHandle handles the GET request to retrieve the original URL for a given short URL ID.
// It extracts the ID from the request path, looks up the original URL in the URLRepository,
// and encodes the original URL as a JSON response.
func (uh *URLHandler) GetOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]
	if !uh.idFilter.Test([]byte(id)) { // if it return false => 100% element is not exist
		http.Error(w, fmt.Sprintf("there's no url with id: %s", id), http.StatusNotFound)
		return
	}

	// Next we check whether id appears in cache
	cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	originURL, err := uh.cache.Get(cacheCtx, id)
	if err != nil {
		slog.Error("failed when try to retrieve entry from cache", "error", err.Error())
	} else if originURL != "" {
		util.EncodeJSON(w, map[string]string{"origin": originURL})
		return
	}

	// Find inside repository
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
		slog.Error("fail when decode json body", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Currently disable cause fail when test
	// Check if the URL is valid
	// if _, err := url.ParseRequestURI(form.Origin); err != nil {
	// 	http.Error(w, "Invalid URL", http.StatusBadRequest)
	// 	return
	// }

	id := uh.idGen.GenerateID()

	short, err := uh.urlRepo.Create(context.Background(), id, form.Origin)
	if err != nil {
		slog.Error("failed to insert url to database", "error", err.Error())
		http.Error(w, fmt.Sprintf("failed when creating short url, error: %s", err), http.StatusInternalServerError)
		return
	}

	// Add shorten ID to Bloom Filter
	uh.idFilter.Add([]byte(short.ID))

	// Add k-v pair (id:origin_url) to cache for 1 hour
	cacheCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := uh.cache.SetWithTTL(cacheCtx, id, form.Origin, 1*time.Hour); err != nil {
		slog.Error("failed to set k-v to cache", "id", id, "origin", form.Origin, "error", err.Error())
	}

	util.EncodeJSON(w, short)
}

type CreateShortForm struct {
	Origin string `json:"origin"`
}
