package storage

import (
	"errors"
)

// Image contains metadata about an image
type Image struct { // TODO: Old api returned width/height for native size and ratios, do that?
	ID     string `json:"id"`
	Author string `json:"author"`
	URL    string `json:"url"`
}

// Provider is an interface for listing and retrieving images
type Provider interface {
	Get(id string) ([]byte, error)
	GetRandom() (id string, err error)
	List() ([]Image, error)
}

// Errors
var (
	ErrNotFound = errors.New("Image does not exist")
)
