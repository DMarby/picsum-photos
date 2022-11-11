package memory_test

import (
	"context"
	"testing"

	"github.com/DMarby/picsum-photos/internal/cache"
	"github.com/DMarby/picsum-photos/internal/cache/memory"
)

func TestMemory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	provider := memory.New()

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
}
