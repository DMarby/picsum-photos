package api

import (
	"expvar"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/params"
	"github.com/gorilla/mux"
	"github.com/twmb/murmur3"
)

var (
	imageRequests          = expvar.NewMap("counter_labelmap_dimensions_image_requests_dimension")
	imageRequestsBlur      = expvar.NewInt("image_requests_blur")
	imageRequestsGrayscale = expvar.NewInt("image_requests_grayscale")
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
	image, err := a.Database.GetRandom(r.Context())
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

	image, handlerErr := a.getImageFromSeed(r, imageSeed)
	if handlerErr != nil {
		return handlerErr
	}

	// Validate the params and redirect to the image service
	return a.validateAndRedirect(w, r, p, image)
}

func (a *API) getImage(r *http.Request, imageID string) (*database.Image, *handler.Error) {
	databaseImage, err := a.Database.Get(r.Context(), imageID)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, &handler.Error{Message: err.Error(), Code: http.StatusNotFound}
		}

		a.logError(r, "error getting image from database", err)
		return nil, handler.InternalServerError()
	}

	return databaseImage, nil
}

func (a *API) getImageFromSeed(r *http.Request, imageSeed string) (*database.Image, *handler.Error) {
	// Hash the input using murmur3
	murmurHash := murmur3.StringSum64(imageSeed)

	// Get a random image by the hash
	image, err := a.Database.GetRandomWithSeed(r.Context(), int64(murmurHash))
	if err != nil {
		a.logError(r, "error getting random image from database", err)
		return nil, handler.InternalServerError()
	}

	return image, nil
}

func (a *API) validateAndRedirect(w http.ResponseWriter, r *http.Request, p *params.Params, image *database.Image) *handler.Error {
	if err := validateImageParams(p); err != nil {
		return handler.BadRequest(err.Error())
	}

	width, height := getImageDimensions(p, image)

	w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
	w.Header()["Content-Type"] = nil

	path := fmt.Sprintf("/id/%s/%d/%d%s", image.ID, width, height, p.Extension)
	query := url.Values{}

	if p.Blur {
		query.Add("blur", strconv.Itoa(p.BlurAmount))
		imageRequestsBlur.Add(1)
	}

	if p.Grayscale {
		query.Add("grayscale", "")
		imageRequestsGrayscale.Add(1)
	}

	url, err := params.HMAC(a.HMAC, path, query)
	if err != nil {
		return handler.InternalServerError()
	}

	imageRequests.Add(fmt.Sprintf("%0.f", math.Max(math.Round(float64(width)/500)*500, math.Round(float64(height)/500)*500)), 1)

	http.Redirect(w, r, fmt.Sprintf("%s%s", a.ImageServiceURL, url), http.StatusFound)

	return nil
}
