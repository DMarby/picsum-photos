package memory_test

import (
	"testing"

	"github.com/DMarby/picsum-photos/cache"
	"github.com/DMarby/picsum-photos/cache/memory"
)

func TestMemory(t *testing.T) {
	provider := memory.New()

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
}
