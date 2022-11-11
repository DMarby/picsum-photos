package redis

import (
	"context"

	"github.com/DMarby/picsum-photos/internal/cache"
	"github.com/DMarby/picsum-photos/internal/tracing"
	"github.com/mediocregopher/radix/v4"
)

// Provider implements a redis cache
type Provider struct {
	client radix.Client
	tracer *tracing.Tracer
}

// New returns a new Provider instance
func New(ctx context.Context, tracer *tracing.Tracer, address string, poolSize int) (*Provider, error) {
	cfg := radix.PoolConfig{
		Size: poolSize,
	}

	client, err := cfg.New(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	return &Provider{
		client: client,
		tracer: tracer,
	}, nil
}

// Get returns an object from the cache if it exists
func (p *Provider) Get(ctx context.Context, key string) (data []byte, err error) {
	ctx, span := p.tracer.Start(ctx, "redis.Get")
	defer span.End()

	mn := radix.Maybe{Rcv: &data}
	err = p.client.Do(ctx, radix.Cmd(&mn, "GET", key))
	if err != nil {
		return nil, err
	}

	if mn.Null {
		return nil, cache.ErrNotFound
	}

	return
}

// Set adds an object to the cache
func (p *Provider) Set(ctx context.Context, key string, data []byte) (err error) {
	ctx, span := p.tracer.Start(ctx, "redis.Set")
	defer span.End()

	return p.client.Do(ctx, radix.FlatCmd(nil, "SET", key, data))
}

// Shutdown shuts down the cache
func (p *Provider) Shutdown() {
	p.client.Close()
}
