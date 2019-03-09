package file

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/DMarby/picsum-photos/database"
)

// Provider implements a file-based image storage
type Provider struct {
	path   string
	images []database.Image
	random *rand.Rand
}

// New returns a new Provider instance
func New(path string) (*Provider, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var images []database.Image
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

func (p *Provider) getImage(id string) (*database.Image, error) {
	for _, image := range p.images {
		if image.ID == id {
			return &image, nil
		}
	}

	return nil, database.ErrNotFound
}

// Get returns the image data for an image id
func (p *Provider) Get(id string) (*database.Image, error) {
	image, err := p.getImage(id)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// GetRandom returns a random image
func (p *Provider) GetRandom() (id string, err error) {
	return p.images[p.random.Intn(len(p.images))].ID, nil
}

// List returns a list of all the images
func (p *Provider) List() ([]database.Image, error) {
	return p.images, nil
}
