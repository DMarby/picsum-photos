package database

import (
	"context"
	"errors"
)

// Image contains metadata about an image
type Image struct {
	ID     string `json:"id"`
	Author string `json:"author"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
}

// Provider is an interface for listing and retrieving images
type Provider interface {
	Get(ctx context.Context, id string) (i *Image, err error)
	GetRandom(ctx context.Context) (i *Image, err error)
	GetRandomWithSeed(ctx context.Context, seed int64) (i *Image, err error)
	ListAll(ctx context.Context) ([]Image, error)
	List(ctx context.Context, offset, limit int) ([]Image, error)

	Wait(ctx context.Context) error
	Migrate(migrationsURL string) error
	Shutdown()
}

// Errors
var (
	ErrNotFound = errors.New("Image does not exist")
)
