//go:build integration
// +build integration

package redis_test

import (
	"testing"

	"github.com/DMarby/picsum-photos/internal/cache"
	"github.com/DMarby/picsum-photos/internal/cache/redis"
	"github.com/mediocregopher/radix/v3"
)

const (
	address  = "127.0.0.1:6380"
	poolSize = 10
)

func TestRedis(t *testing.T) {
	provider, err := redis.New(address, poolSize)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := radix.NewPool("tcp", address, poolSize)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	t.Run("get item", func(t *testing.T) {
		// Add item to the cache
		provider.Set("foo", []byte("bar"))

		// Get item from the cache
		data, err := provider.Get("foo")
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != "bar" {
			t.Fatal("wrong data")
		}
	})

	t.Run("get nonexistant item", func(t *testing.T) {
		_, err := provider.Get("notfound")
		if err == nil {
			t.Fatal("no error")
		}

		if err != cache.ErrNotFound {
			t.Fatalf("wrong error %s", err)
		}
	})

	t.Run("get error", func(t *testing.T) {
		provider.Shutdown()
		_, err := provider.Get("notfound")
		if err == nil {
			t.Fatal("no error")
		}
	})

	// Clean up
	pool.Do(radix.Cmd(nil, "FLUSHALL"))
}

func TestNew(t *testing.T) {
	_, err := redis.New("", 10)
	if err == nil {
		t.Fatal("no error")
	}
}
