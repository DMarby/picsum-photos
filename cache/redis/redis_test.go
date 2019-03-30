package redis_test

import (
	"testing"

	"github.com/DMarby/picsum-photos/cache"
	"github.com/DMarby/picsum-photos/cache/redis"
)

func TestRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	provider, err := redis.New("127.0.0.1:6379", 10)
	if err != nil {
		t.Fatal(err)
	}

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
}

func TestNew(t *testing.T) {
	_, err := redis.New("", 10)
	if err == nil {
		t.Fatal("no error")
	}
}
