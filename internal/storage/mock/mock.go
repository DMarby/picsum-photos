package mock

import (
	"context"
)

// Provider implements a mock image storage
type Provider struct {
}

// Get returns the image data for an image id
func (p *Provider) Get(ctx context.Context, id string) ([]byte, error) {
	return []byte("foo"), nil
}
