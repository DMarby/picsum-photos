package memory

import "github.com/DMarby/picsum-photos/cache"

// Provider implements a simple in-memory cache
type Provider struct {
	cache map[string][]byte
}

// New returns a new Provider instance
func New() *Provider {
	return &Provider{
		cache: make(map[string][]byte),
	}
}

// Get returns an object from the cache if it exists
func (p *Provider) Get(key string) (data []byte, err error) {
	data, exists := p.cache[key]
	if !exists {
		return nil, cache.ErrNotFound
	}

	return data, nil
}

// Set returns an object from the cache if it exists
func (p *Provider) Set(key string, data []byte) (err error) {
	p.cache[key] = data
	return nil
}
