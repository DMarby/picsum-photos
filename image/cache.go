package image

import (
	"github.com/DMarby/picsum-photos/cache"
	"github.com/DMarby/picsum-photos/storage"
)

// Cache is an image cache
type Cache = cache.Auto

// NewCache instantiates a new cache
func NewCache(cacheProvider cache.Provider, storageProvider storage.Provider) *Cache {
	return &Cache{
		Provider: cacheProvider,
		Loader: func(key string) (data []byte, err error) {
			return storageProvider.Get(key)
		},
	}
}
