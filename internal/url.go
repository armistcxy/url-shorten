package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
)

type ShortURL struct {
	ID        string    `json:"id"`
	Origin    string    `json:"origin"`
	CreatedAt time.Time `json:"created_at"`
}

type URLRepository interface {
	Create(ctx context.Context, url string) (*ShortURL, error)
	Get(ctx context.Context, id string) (string, error)
}

// This will deal with 2 end points
// /short/:id GET => Return original url
// /create?url= POST => Return short url
type URLHandler struct {
	repo     URLRepository
	idFilter *bloom.BloomFilter
}

func NewURLHandler(repo URLRepository) *URLHandler {
	return &URLHandler{
		repo:     repo,
		idFilter: bloom.NewWithEstimates(1_000_000, 0.01),
	}
}

// GetOriginURLHandle handles the GET request to retrieve the original URL for a given short URL ID.
// It extracts the ID from the request path, looks up the original URL in the URLRepository,
// and encodes the original URL as a JSON response.
func (uh *URLHandler) GetOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if len(id) != 6 {
		http.Error(w, "id length must be equal 6", http.StatusBadRequest)
		return
	}
	if !uh.idFilter.Test([]byte(id)) { // if it return false => 100% element is not exist
		http.Error(w, fmt.Sprintf("there's no url with id: %s", id), http.StatusNotFound)
		return
	}
	originURL, err := uh.repo.Get(context.Background(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("fail to retrive origin url, error: %s", err), http.StatusInternalServerError)
		return
	}
	EncodeJSON(w, map[string]string{"origin": originURL})
}

// CreateShortURLHandle handles the POST request to create a new short URL.
// It extracts the original URL from the request, creates a new short URL using the URLRepository,
// and encodes the short URL as a JSON response.
func (uh *URLHandler) CreateShortURLHandle(w http.ResponseWriter, r *http.Request) {
	form := CreateShortForm{}
	if err := DecodeJSON(r, &form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		slog.Error("fail when decode json body", "error", err.Error())
		return
	}

	// Check if the URL is valid
	if _, err := url.ParseRequestURI(form.Origin); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	short, err := uh.repo.Create(context.Background(), form.Origin)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed when creating short url, error: %s", err), http.StatusInternalServerError)
		return
	}
	uh.idFilter.Add([]byte(short.ID))
	EncodeJSON(w, short)
}

type CreateShortForm struct {
	Origin string `json:"origin"`
}
