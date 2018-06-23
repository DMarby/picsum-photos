package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/DMarby/picsum-photos/storage"
)

// Provider implements a file-based image storage
type Provider struct {
	path   string
	images []storage.Image
	random *rand.Rand
}

// New returns a new Provider instance
func New(path string) (*Provider, error) {
	data, err := ioutil.ReadFile(filepath.Join(path, "metadata.json"))
	if err != nil {
		return nil, err
	}

	var images []storage.Image
	err = json.Unmarshal(data, &images)
	if err != nil {
		return nil, err
	}

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	return &Provider{
		path,
		images,
		random,
	}, nil
}

func (p *Provider) getImage(id string) (*storage.Image, error) {
	for _, image := range p.images {
		if image.ID == id {
			return &image, nil
		}
	}

	return nil, storage.ErrNotFound
}

// Get returns the image data for an image id
func (p *Provider) Get(id string) ([]byte, error) {
	image, err := p.getImage(id)
	if err != nil {
		return nil, err
	}

	imageData, err := ioutil.ReadFile(filepath.Join(p.path, fmt.Sprintf("%s.jpg", image.ID)))
	if err != nil {
		return nil, err
	}

	return imageData, nil
}

// GetRandom returns a random image
func (p *Provider) GetRandom() (id string, err error) {
	return p.images[p.random.Intn(len(p.images))].ID, nil
}

// List returns a list of all the images
func (p *Provider) List() ([]storage.Image, error) {
	return p.images, nil
}
