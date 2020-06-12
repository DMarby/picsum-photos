package params

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Errors
var (
	ErrInvalidSize          = fmt.Errorf("Invalid size")
	ErrInvalidFileExtension = fmt.Errorf("Invalid file extension")
)

const defaultBlurAmount = 5

// Params contains all the parameters for a request
type Params struct {
	Width      int
	Height     int
	Blur       bool
	BlurAmount int
	Grayscale  bool
	Extension  string
}

// GetParams parses and returns all the path and query parameters
func GetParams(r *http.Request) (*Params, error) {
	// Get and validate the width and height from the path parameters
	width, height, err := getSize(r)
	if err != nil {
		return nil, err
	}

	// Get the optional file extension from the path parameters
	extension, err := getFileExtension(r)
	if err != nil {
		return nil, err
	}

	// Get and validate the query parameters for grayscale and blur
	grayscale, blur, blurAmount := getQueryParams(r)

	params := &Params{
		Width:      width,
		Height:     height,
		Blur:       blur,
		BlurAmount: blurAmount,
		Grayscale:  grayscale,
		Extension:  extension,
	}

	return params, nil
}

// getSize gets the image size from the size or the width/height path params, and validates it
func getSize(r *http.Request) (width int, height int, err error) {
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

// intParam tries to get a param and convert it to an Integer
func intParam(r *http.Request, name string) (int, bool) {
	vars := mux.Vars(r)

	if val, ok := vars[name]; ok {
		val, err := strconv.Atoi(val)
		return val, err == nil
	}

	return -1, false
}

// getFileExtension gets the file extension (if present) from the path params, and validates it
func getFileExtension(r *http.Request) (extension string, err error) {
	vars := mux.Vars(r)

	// We only allow the .jpg and .webp extensions, as we only serve jpg and webp images
	// We normalize having no extension since it's an optional path param
	val := strings.ToLower(vars["extension"])

	if val == "" {
		val = ".jpg"
	}

	if val != ".jpg" && val != ".webp" {
		return "", ErrInvalidFileExtension
	}

	return val, nil
}

// getQueryParams returns whether the grayscale and blur queryparams are present
func getQueryParams(r *http.Request) (grayscale bool, blur bool, blurAmount int) {
	if _, ok := r.URL.Query()["grayscale"]; ok {
		grayscale = true
	}

	if _, ok := r.URL.Query()["blur"]; ok {
		blur = true
		blurAmount = defaultBlurAmount

		if val, err := strconv.Atoi(r.URL.Query().Get("blur")); err == nil {
			blurAmount = val
			return
		}
	}

	return
}
