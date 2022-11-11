package memory

import (
	"context"
	"sync"

	"github.com/DMarby/picsum-photos/internal/cache"
)

// Provider implements a simple in-memory cache
type Provider struct {
	cache map[string][]byte
	mutex sync.RWMutex
}

// New returns a new Provider instance
func New() *Provider {
	return &Provider{
		cache: make(map[string][]byte),
	}
}

// Get returns an object from the cache if it exists
func (p *Provider) Get(ctx context.Context, key string) (data []byte, err error) {
	p.mutex.RLock()
	data, exists := p.cache[key]
	p.mutex.RUnlock()

	if !exists {
		return nil, cache.ErrNotFound
	}

	return data, nil
}

// Set adds an object to the cache
func (p *Provider) Set(ctx context.Context, key string, data []byte) (err error) {
	p.mutex.Lock()
	p.cache[key] = data
	p.mutex.Unlock()

	return nil
}

// Shutdown shuts down the cache
func (p *Provider) Shutdown() {}
