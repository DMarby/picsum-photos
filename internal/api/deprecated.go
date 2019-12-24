package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DMarby/picsum-photos/internal/database"

	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/params"
)

// DeprecatedImage contains info about an image, in the old deprecated /list style
type DeprecatedImage struct {
	Format    string `json:"format"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Filename  string `json:"filename"`
	ID        int    `json:"id"`
	Author    string `json:"author"`
	AuthorURL string `json:"author_url"`
	PostURL   string `json:"post_url"`
}

func (a *API) deprecatedListHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	list, err := a.Database.ListAll()
	if err != nil {
		a.logError(r, "error getting image list from database", err)
		return handler.InternalServerError()
	}

	var images []DeprecatedImage
	for _, image := range list {
		numericID, err := strconv.Atoi(image.ID)
		if err != nil {
			continue
		}

		images = append(images, DeprecatedImage{
			Format:    "jpeg",
			Width:     image.Width,
			Height:    image.Height,
			Filename:  fmt.Sprintf("%s.jpeg", image.ID),
			ID:        numericID,
			Author:    image.Author,
			AuthorURL: image.URL,
			PostURL:   image.URL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(images); err != nil {
		a.logError(r, "error encoding deprecate image list", err)
		return handler.InternalServerError()
	}

	return nil
}

// Handles deprecated image routes
func (a *API) deprecatedImageHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	// Get the params
	p, err := params.GetParams(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	var image *database.Image

	// Look for the deprecated ?image query parameter
	if id := r.URL.Query().Get("image"); id != "" {
		var handlerErr *handler.Error
		image, handlerErr = a.getImage(r, id)
		if handlerErr != nil {
			return handlerErr
		}
	} else {
		image, err = a.Database.GetRandom()
		if err != nil {
			a.logError(r, "error getting random image from database", err)
			return handler.InternalServerError()
		}
	}

	// Set grayscale to true as this is the deprecated /g/ endpoint
	p.Grayscale = true

	return a.validateAndRedirect(w, r, p, image)
}

// deprecatedParams is a handler to handle deprecated query params for regular routes
func (a *API) deprecatedParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Look for the deprecated ?image query parameter
		if id := r.URL.Query().Get("image"); id != "" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

			p, err := params.GetParams(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			image, handlerErr := a.getImage(r, id)
			if handlerErr != nil {
				http.Error(w, handlerErr.Message, handlerErr.Code)
				return
			}

			handlerErr = a.validateAndRedirect(w, r, p, image)
			if handlerErr != nil {
				http.Error(w, handlerErr.Message, handlerErr.Code)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}
