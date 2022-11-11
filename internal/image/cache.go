package image

import (
	"context"

	"github.com/DMarby/picsum-photos/internal/cache"
	"github.com/DMarby/picsum-photos/internal/storage"
	"github.com/DMarby/picsum-photos/internal/tracing"
)

// Cache is an image cache
type Cache = cache.Auto

// NewCache instantiates a new cache
func NewCache(tracer *tracing.Tracer, cacheProvider cache.Provider, storageProvider storage.Provider) *Cache {
	return &Cache{
		Tracer:   tracer,
		Provider: cacheProvider,
		Loader: func(ctx context.Context, key string) (data []byte, err error) {
			ctx, span := tracer.Start(ctx, "image.Cache.Loader")
			defer span.End()

			return storageProvider.Get(ctx, key)
		},
	}
}
