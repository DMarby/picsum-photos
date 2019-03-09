package mock

import (
	"fmt"
)

// Provider implements a mock image storage
type Provider struct {
}

// Get returns the image data for an image id
func (p *Provider) Get(id string) ([]byte, error) {
	return nil, fmt.Errorf("get error")
}
