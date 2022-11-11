//go:build integration
// +build integration

package redis_test

import (
	"context"
	"testing"

	"github.com/DMarby/picsum-photos/internal/cache"
	"github.com/DMarby/picsum-photos/internal/cache/redis"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/tracing/test"
	"github.com/mediocregopher/radix/v4"
	"go.uber.org/zap"
)

const (
	address  = "127.0.0.1:6380"
	poolSize = 10
)

func TestRedis(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.New(zap.ErrorLevel)
	defer log.Sync()

	tracer := test.Tracer(log)

	provider, err := redis.New(ctx, tracer, address, poolSize)
	if err != nil {
		t.Fatal(err)
	}

	cfg := radix.PoolConfig{}
	client, err := cfg.New(ctx, "tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	t.Run("get item", func(t *testing.T) {
		// Add item to the cache
		provider.Set(ctx, "foo", []byte("bar"))

		// Get item from the cache
		data, err := provider.Get(ctx, "foo")
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != "bar" {
			t.Fatal("wrong data")
		}
	})

	t.Run("get nonexistant item", func(t *testing.T) {
		_, err := provider.Get(ctx, "notfound")
		if err == nil {
			t.Fatal("no error")
		}

		if err != cache.ErrNotFound {
			t.Fatalf("wrong error %s", err)
		}
	})

	t.Run("get error", func(t *testing.T) {
		provider.Shutdown()
		_, err := provider.Get(ctx, "notfound")
		if err == nil {
			t.Fatal("no error")
		}
	})

	// Clean up
	client.Do(ctx, radix.Cmd(nil, "FLUSHALL"))
}

func TestNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := redis.New(ctx, nil, "", 10)
	if err == nil {
		t.Fatal("no error")
	}
}
