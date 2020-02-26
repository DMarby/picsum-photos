package api

import (
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/params"
	"github.com/gorilla/mux"
	"github.com/twmb/murmur3"
)

func (a *API) imageRedirectHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	// Get the path and query parameters
	p, err := params.GetParams(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	// Get the image from the database
	vars := mux.Vars(r)
	imageID := vars["id"]
	image, handlerErr := a.getImage(r, imageID)
	if handlerErr != nil {
		return handlerErr
	}

	// Validate the params and redirect to the image service
	return a.validateAndRedirect(w, r, p, image)
}

func (a *API) randomImageRedirectHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	// Get the path and query parameters
	p, err := params.GetParams(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	// Get a random image
	image, err := a.Database.GetRandom()
	if err != nil {
		a.logError(r, "error getting random image from database", err)
		return handler.InternalServerError()
	}

	// Validate the params and redirect to the image service
	return a.validateAndRedirect(w, r, p, image)
}

func (a *API) seedImageRedirectHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	// Get the path and query parameters
	p, err := params.GetParams(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	// Get the image seed
	vars := mux.Vars(r)
	imageSeed := vars["seed"]

	// Hash the input using murmur3
	murmurHash := murmur3.StringSum64(imageSeed)

	// Get a random image by the hash
	image, err := a.Database.GetRandomWithSeed(int64(murmurHash))
	if err != nil {
		a.logError(r, "error getting random image from database", err)
		return handler.InternalServerError()
	}

	// Validate the params and redirect to the image service
	return a.validateAndRedirect(w, r, p, image)
}

func (a *API) getImage(r *http.Request, imageID string) (*database.Image, *handler.Error) {
	databaseImage, err := a.Database.Get(imageID)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, &handler.Error{Message: err.Error(), Code: http.StatusNotFound}
		}

		a.logError(r, "error getting image from database", err)
		return nil, handler.InternalServerError()
	}

	return databaseImage, nil
}

func (a *API) validateAndRedirect(w http.ResponseWriter, r *http.Request, p *params.Params, image *database.Image) *handler.Error {
	if err := p.Validate(image); err != nil {
		return handler.BadRequest(err.Error())
	}

	width, height := p.Dimensions(image)

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header()["Content-Type"] = nil
	http.Redirect(w, r, fmt.Sprintf("%s/id/%s/%d/%d%s%s", a.ImageServiceURL, image.ID, width, height, p.Extension, params.BuildQuery(p.Grayscale, p.Blur, p.BlurAmount)), http.StatusFound)

	return nil
}
