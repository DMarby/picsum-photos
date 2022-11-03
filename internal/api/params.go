package api

import (
	"fmt"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/params"
)

// Errors
var (
	ErrInvalidBlurAmount = fmt.Errorf("Invalid blur amount")
)

const (
	minBlurAmount = 1
	maxBlurAmount = 10
	maxImageSize  = 5000 // The max allowed image width/height that can be requested
)

func validateImageParams(p *params.Params) error {
	if p.Width > maxImageSize {
		return params.ErrInvalidSize
	}

	if p.Height > maxImageSize {
		return params.ErrInvalidSize
	}

	if p.Blur && p.BlurAmount < minBlurAmount {
		return ErrInvalidBlurAmount
	}

	if p.Blur && p.BlurAmount > maxBlurAmount {
		return ErrInvalidBlurAmount
	}

	return nil
}

func getImageDimensions(p *params.Params, databaseImage *database.Image) (width, height int) {
	// Default to the image width/height if 0 is passed
	width = p.Width
	height = p.Height

	if width == 0 {
		width = databaseImage.Width
	}

	if height == 0 {
		height = databaseImage.Height
	}

	return
}
