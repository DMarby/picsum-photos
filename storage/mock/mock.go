package mock

import (
	"fmt"

	"github.com/DMarby/picsum-photos/storage"
)

// Provider implements a mock image storage
type Provider struct {
}

// Get returns the image data for an image id
func (p *Provider) Get(id string) ([]byte, error) {
	return nil, fmt.Errorf("get error")
}

// GetRandom returns a random image
func (p *Provider) GetRandom() (id string, err error) {
	return "", fmt.Errorf("random error")
}

// List returns a list of all the images
func (p *Provider) List() ([]storage.Image, error) {
	return nil, fmt.Errorf("list error")
}
