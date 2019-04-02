package redis

import (
	"github.com/DMarby/picsum-photos/cache"
	"github.com/mediocregopher/radix/v3"
)

// Provider implements a redis cache
type Provider struct {
	pool *radix.Pool
}

// New returns a new Provider instance
func New(address string, poolSize int) (*Provider, error) {
	// Use the default pool, which has a 10 second timeout
	pool, err := radix.NewPool("tcp", address, poolSize)
	if err != nil {
		return nil, err
	}

	return &Provider{
		pool: pool,
	}, nil
}

// Get returns an object from the cache if it exists
func (p *Provider) Get(key string) (data []byte, err error) {
	mn := radix.MaybeNil{Rcv: &data}
	err = p.pool.Do(radix.Cmd(&mn, "GET", key))
	if err != nil {
		return nil, err
	}

	if mn.Nil {
		return nil, cache.ErrNotFound
	}

	return
}

// Set adds an object to the cache
func (p *Provider) Set(key string, data []byte) (err error) {
	return p.pool.Do(radix.FlatCmd(nil, "SET", key, data))
}

// Shutdown shuts down the cache
func (p *Provider) Shutdown() {
	p.pool.Close()
}
