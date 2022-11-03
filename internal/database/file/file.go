package file

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/DMarby/picsum-photos/internal/database"
)

// Provider implements a file-based image storage
type Provider struct {
	images       []database.Image
	sortedImages []database.Image

	random *rand.Rand
	mu     sync.Mutex
}

// New returns a new Provider instance
func New(path string) (*Provider, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var images []database.Image
	err = json.Unmarshal(data, &images)
	if err != nil {
		return nil, err
	}

	sortedImages := make([]database.Image, len(images))
	copy(sortedImages, images)
	sort.Slice(sortedImages, func(i, j int) bool {
		ii, _ := strconv.Atoi(sortedImages[i].ID)
		jj, _ := strconv.Atoi(sortedImages[j].ID)
		return ii < jj
	})

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	return &Provider{
		images:       images,
		sortedImages: sortedImages,
		random:       random,
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
func (p *Provider) Get(ctx context.Context, id string) (i *database.Image, err error) {
	image, err := p.getImage(id)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// GetRandom returns a random image ID
func (p *Provider) GetRandom(ctx context.Context) (i *database.Image, err error) {
	p.mu.Lock()
	image := &p.images[p.random.Intn(len(p.images))]
	p.mu.Unlock()
	return image, nil
}

// GetRandomWithSeed returns a random image ID based on the given seed
func (p *Provider) GetRandomWithSeed(ctx context.Context, seed int64) (i *database.Image, err error) {
	source := rand.NewSource(seed)
	random := rand.New(source)

	return &p.images[random.Intn(len(p.images))], nil
}

// ListAll returns a list of all the images
func (p *Provider) ListAll(ctx context.Context) ([]database.Image, error) {
	return p.sortedImages, nil
}

// List returns a list of all the images with an offset/limit
func (p *Provider) List(ctx context.Context, offset, limit int) ([]database.Image, error) {
	images := len(p.sortedImages)
	if offset > images {
		offset = images
	}

	limit = offset + limit
	if limit > images {
		limit = images
	}

	return p.sortedImages[offset:limit], nil
}
