package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Provider implements a file-based image storage
type Provider struct {
	path string
}

// New returns a new Provider instance
func New(path string) (*Provider, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	return &Provider{
		path,
	}, nil
}

// Get returns the image data for an image id
func (p *Provider) Get(id string) ([]byte, error) {
	imageData, err := ioutil.ReadFile(filepath.Join(p.path, fmt.Sprintf("%s.jpg", id)))
	if err != nil {
		return nil, err
	}

	return imageData, nil
}
