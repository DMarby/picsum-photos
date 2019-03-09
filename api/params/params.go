package params

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DMarby/picsum-photos/database"
	"github.com/gorilla/mux"
)

// intParam tries to get a param and convert it to an Integer
func intParam(r *http.Request, name string) (int, bool) {
	vars := mux.Vars(r)

	if val, ok := vars[name]; ok {
		val, err := strconv.Atoi(val)
		return val, err == nil
	}

	return -1, false
}

// Errors
var (
	ErrInvalidSize       = fmt.Errorf("Invalid size")
	ErrInvalidBlurAmount = fmt.Errorf("Invalid blur amount")
)

// ValidateParams checks that the size is within the allowed limit
func ValidateParams(maxImageSize int, image *database.Image, width int, height int, blur bool, blurAmount int) error {
	if width < 1 || height < 1 {
		return ErrInvalidSize
	}

	if width > maxImageSize && width != image.Width {
		return ErrInvalidSize
	}

	if height > maxImageSize && height != image.Height {
		return ErrInvalidSize
	}

	if blur && blurAmount < 1 {
		return ErrInvalidBlurAmount
	}

	if blur && blurAmount > 10 {
		return ErrInvalidBlurAmount
	}

	return nil
}

// GetSize gets the image size from the size or the width/height path params, and validates it
func GetSize(r *http.Request) (width int, height int, err error) {
	// Check for the size parameter first
	if size, ok := intParam(r, "size"); ok {
		width, height = size, size
	} else {
		// If size doesn't exist, check for width/height
		width, ok = intParam(r, "width")
		if !ok {
			return -1, -1, ErrInvalidSize
		}

		height, ok = intParam(r, "height")
		if !ok {
			return -1, -1, ErrInvalidSize
		}
	}

	return
}

// GetQueryParams returns whether the grayscale and blur queryparams are present
func GetQueryParams(r *http.Request) (grayscale bool, blur bool, blurAmount int) {
	if _, ok := r.URL.Query()["grayscale"]; ok {
		grayscale = true
	}

	if _, ok := r.URL.Query()["blur"]; ok {
		blur = true
		blurAmount = 5

		if val, err := strconv.Atoi(r.URL.Query().Get("blur")); err == nil {
			blurAmount = val
			return
		}
	}

	return
}

// Utilities for building a URL with query params

// addParam adds a query parameter to a byte buffer
func addParam(buf *bytes.Buffer, param string) {
	if buf.Len() > 0 {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}

	buf.WriteString(param)
}

// BuildQuery builds query parameters for the given arguments
func BuildQuery(grayscale bool, blur bool, blurAmount int) string {
	if !grayscale && !blur {
		return ""
	}

	var buf bytes.Buffer

	if grayscale {
		addParam(&buf, "grayscale")
	}

	if blur {
		addParam(&buf, fmt.Sprintf("blur=%d", blurAmount))
	}

	return buf.String()
}
