package imageapi

import (
	"net/http"
	"time"

	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/hmac"

	"github.com/DMarby/picsum-photos/internal/health"
	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/gorilla/mux"
)

// API is a http api
type API struct {
	ImageProcessor image.Processor
	HealthChecker  *health.Checker
	Log            *logger.Logger
	HandlerTimeout time.Duration
	HMAC           *hmac.HMAC
}

// Utility methods for logging
func (a *API) logError(r *http.Request, message string, err error) {
	a.Log.Errorw(message, handler.LogFields(r, "error", err)...)
}

// Router returns a http router
func (a *API) Router() http.Handler {
	router := mux.NewRouter()

	router.NotFoundHandler = handler.Handler(a.notFoundHandler)

	// Redirect trailing slashes
	router.StrictSlash(true)

	// Healthcheck
	router.Handle("/health", handler.Health(a.HealthChecker)).Methods("GET")

	// Image by ID routes
	router.Handle("/id/{id}/{width:[0-9]+}/{height:[0-9]+}{extension:\\..*}", handler.Handler(a.imageHandler)).Methods("GET")

	// Query parameters:
	// ?grayscale - Grayscale the image
	// ?blur - Blur the image
	// ?blur={amount} - Blur the image by {amount}

	// ?hmac - HMAC signature of the path and URL parameters

	// Set up handlers for adding a request id, handling panics, request logging, setting CORS headers, and handler execution timeout
	return handler.AddRequestID(handler.Recovery(a.Log, handler.Logger(a.Log, handler.CORS([]string{"Picsum-ID"}, http.TimeoutHandler(router, a.HandlerTimeout, "Something went wrong. Timed out.")))))
}

// Handle not found errors
var notFoundError = &handler.Error{
	Message: "page not found",
	Code:    http.StatusNotFound,
}

func (a *API) notFoundHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	return notFoundError
}
