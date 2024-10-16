package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

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
}

func NewURLHandler(urlRepo domain.URLRepository, idGen domain.IDGenerator) *URLHandler {
	return &URLHandler{
		urlRepo:  urlRepo,
		idGen:    idGen,
		idFilter: bloom.NewWithEstimates(1_000_000, 0.01),
	}
}

// GetOriginURLHandle handles the GET request to retrieve the original URL for a given short URL ID.
// It extracts the ID from the request path, looks up the original URL in the URLRepository,
// and encodes the original URL as a JSON response.
func (uh *URLHandler) GetOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !uh.idFilter.Test([]byte(id)) { // if it return false => 100% element is not exist
		http.Error(w, fmt.Sprintf("there's no url with id: %s", id), http.StatusNotFound)
		return
	}
	originURL, err := uh.urlRepo.Get(context.Background(), id)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		slog.Error("fail when decode json body", "error", err.Error())
		return
	}

	// Check if the URL is valid
	if _, err := url.ParseRequestURI(form.Origin); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id := uh.idGen.GenerateID()

	short, err := uh.urlRepo.Create(context.Background(), id, form.Origin)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed when creating short url, error: %s", err), http.StatusInternalServerError)
		return
	}
	uh.idFilter.Add([]byte(short.ID))
	util.EncodeJSON(w, short)
}

type CreateShortForm struct {
	Origin string `json:"origin"`
}
