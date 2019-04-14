package storage

import "context"

// Provider is an interface for retrieving images
type Provider interface {
	Get(ctx context.Context, id string) ([]byte, error)
}
