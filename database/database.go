package database

import (
	"errors"
)

// Image contains metadata about an image
type Image struct {
	ID     string `json:"id"`
	Author string `json:"author"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
}

// Provider is an interface for listing and retrieving images
type Provider interface {
	Get(id string) (*Image, error)
	GetRandom() (id string, err error)
	List() ([]Image, error)
	Shutdown()
}

// Errors
var (
	ErrNotFound = errors.New("Image does not exist")
)
