package domain

import (
	"context"
	"time"
)

type ShortURL struct {
	ID        string    `json:"id"`
	Origin    string    `json:"origin"`
	CreatedAt time.Time `json:"created_at"`
}

type URLRepository interface {
	Create(ctx context.Context, id string, url string) (*ShortURL, error)
	Get(ctx context.Context, id string) (string, error)
}

type IDGenerator interface {
	GenerateID() string
}
