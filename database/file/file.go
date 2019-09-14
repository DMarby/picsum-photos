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

// GetRandom returns a random image ID
func (p *Provider) GetRandom() (id string, err error) {
	return p.images[p.random.Intn(len(p.images))].ID, nil
}

// GetRandomWithSeed returns a random image ID based on the given seed
func (p *Provider) GetRandomWithSeed(seed int64) (id string, err error) {
	source := rand.NewSource(seed)
	random := rand.New(source)

	return p.images[random.Intn(len(p.images))].ID, nil
}

// ListAll returns a list of all the images
func (p *Provider) ListAll() ([]database.Image, error) {
	return p.images, nil
}

// List returns a list of all the images with an offset/limit
func (p *Provider) List(offset, limit int) ([]database.Image, error) {
	images := len(p.images)
	if offset > images {
		offset = images
	}

	limit = offset + limit
	if limit > images {
		limit = images
	}

	return p.images[offset:limit], nil
}

// Shutdown shuts down the database client
func (p *Provider) Shutdown() {}
