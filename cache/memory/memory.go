package memory

import (
	"sync"

	"github.com/DMarby/picsum-photos/cache"
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
func (p *Provider) Get(key string) (data []byte, err error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	data, exists := p.cache[key]
	if !exists {
		return nil, cache.ErrNotFound
	}

	return data, nil
}

// Set adds an object to the cache
func (p *Provider) Set(key string, data []byte) (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.cache[key] = data
	return nil
}

// Shutdown shuts down the cache
func (p *Provider) Shutdown() {}
