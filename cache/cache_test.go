package cache_test

import (
	"fmt"
	"testing"

	"github.com/DMarby/picsum-photos/cache"
)

type mockProvider struct{}

// Get returns an object from the cache if it exists
func (p *mockProvider) Get(key string) (data []byte, err error) {
	if key == "notfound" || key == "notfounderr" || key == "seterror" {
		return nil, cache.ErrNotFound
	}

	if key == "error" {
		return nil, fmt.Errorf("error")
	}

	return []byte("foo"), nil
}

// Set returns an object from the cache if it exists
func (p *mockProvider) Set(key string, data []byte) (err error) {
	if key == "seterror" {
		return fmt.Errorf("seterror")
	}

	return nil
}

var mockLoaderFunc cache.LoaderFunc = func(key string) (data []byte, err error) {
	if key == "notfounderr" {
		return nil, fmt.Errorf("notfounderr")
	}

	return []byte("notfound"), nil
}

func TestAuto(t *testing.T) {
	auto := &cache.Auto{
		Provider: &mockProvider{},
		Loader:   mockLoaderFunc,
	}

	tests := []struct {
		Key           string
		ExpectedError error
	}{
		{"foo", nil},
		{"notfound", nil},
		{"notfounderr", fmt.Errorf("notfounderr")},
		{"seterror", fmt.Errorf("seterror")},
	}

	for _, test := range tests {
		data, err := auto.Get(test.Key)
		if err != nil {
			if test.ExpectedError == nil {
				t.Errorf("%s: %s", test.Key, err)
				continue
			}

			if test.ExpectedError.Error() != err.Error() {
				t.Errorf("%s: wrong error: %s", test.Key, err)
				continue
			}

			continue
		}

		if string(data) != test.Key {
			t.Errorf("%s: wrong data", test.Key)
		}
	}

}
