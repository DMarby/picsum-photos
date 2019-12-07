package storage

import (
	"context"
	"errors"
)

// Provider is an interface for retrieving images
type Provider interface {
	Get(ctx context.Context, id string) ([]byte, error)
}

// Errors
var (
	ErrNotFound = errors.New("Image does not exist")
)
