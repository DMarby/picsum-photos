package api

import (
	"github.com/DMarby/picsum-photos/cache"
	"github.com/DMarby/picsum-photos/storage"
)

// ImageCache is an image cache
type ImageCache = cache.Auto

// NewCache instantiates a new cache
func NewCache(cacheProvider cache.Provider, storageProvider storage.Provider) *ImageCache {
	return &ImageCache{
		Provider: cacheProvider,
		Loader: func(key string) (data []byte, err error) {
			return storageProvider.Get(key)
		},
	}
}
