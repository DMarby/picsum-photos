package cache

import (
	"context"
	"errors"

	"golang.org/x/sync/singleflight"
)

// Provider is an interface for getting and setting cached objects
type Provider interface {
	Get(key string) (data []byte, err error)
	Set(key string, data []byte) (err error)
	Shutdown()
}

// LoaderFunc is a function for loading data into a cache
type LoaderFunc func(ctx context.Context, key string) (data []byte, err error)

// Auto is a cache that automatically attempts to load objects if they don't exist
type Auto struct {
	Provider    Provider
	Loader      LoaderFunc
	lookupGroup singleflight.Group
}

// Get returns an object from the cache if it exists, otherwise it loads it into the cache and returns it
func (a *Auto) Get(ctx context.Context, key string) (data []byte, err error) {
	// Attempt to get the data from the cache
	data, err = a.Provider.Get(key)
	// Exit early if the error is nil as we got data from the cache
	// Or if there's an error indicating that something went wrong
	if err != ErrNotFound {
		return
	}

	// Use singleflight to avoid concurrent requests
	var v interface{}
	v, err, _ = a.lookupGroup.Do(key, func() (interface{}, error) {
		// Get the data
		data, err := a.Loader(ctx, key)
		if err != nil {
			return nil, err
		}

		// Store the data in the cache
		err = a.Provider.Set(key, data)
		if err != nil {
			return nil, err
		}

		return data, nil
	})

	if err != nil {
		return
	}

	data, _ = v.([]byte)
	return
}

// Errors
var (
	ErrNotFound = errors.New("not found in cache")
)
