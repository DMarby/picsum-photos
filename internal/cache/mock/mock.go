package mock

import (
	"context"
	"fmt"

	"github.com/DMarby/picsum-photos/internal/cache"
)

// Provider is a mock cache
type Provider struct{}

// Get returns an object from the cache if it exists
func (p *Provider) Get(ctx context.Context, key string) (data []byte, err error) {
	if key == "notfound" || key == "notfounderr" || key == "seterror" {
		return nil, cache.ErrNotFound
	}

	if key == "error" {
		return nil, fmt.Errorf("error")
	}

	return []byte("foo"), nil
}

// Set adds an object to the cache
func (p *Provider) Set(ctx context.Context, key string, data []byte) (err error) {
	if key == "seterror" {
		return fmt.Errorf("seterror")
	}

	return nil
}

// Shutdown shuts down the cache
func (p *Provider) Shutdown() {}
