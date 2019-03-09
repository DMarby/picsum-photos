package api

import (
	"net/http"

	"github.com/DMarby/picsum-photos/api/handler"
	"github.com/DMarby/picsum-photos/api/middleware"

	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/health"
	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/logger"
	"github.com/DMarby/picsum-photos/storage"
	"github.com/gorilla/mux"
)

// API is a http api
type API struct {
	ImageProcessor image.Processor
	Storage        storage.Provider
	Database       database.Provider
	HealthChecker  *health.Checker
	Log            *logger.Logger
	MaxImageSize   int
	RootURL        string
}

// Utility methods for logging
func (a *API) logError(r *http.Request, message string, err error) {
	a.Log.Errorw(message, logFields(r, "error", err)...)
}

func logFields(r *http.Request, keysAndValues ...interface{}) []interface{} {
	ctx := r.Context()
	id := middleware.GetReqID(ctx)

	return append([]interface{}{"request-id", id}, keysAndValues...)
}

// Handle not found errors
var notFoundError = &handler.Error{
	Message: "page not found",
	Code:    http.StatusNotFound,
}

func (a *API) notFoundHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	return notFoundError
}

// Router returns a http router
func (a *API) Router() http.Handler {
	router := mux.NewRouter()

	router.NotFoundHandler = handler.Handler(a.notFoundHandler)

	// Redirect trailing slashes
	router.StrictSlash(true)

	// Healthcheck
	router.Handle("/health", handler.Handler(a.healthHandler)).Methods("GET")

	// Image list
	router.Handle("/v2/list", handler.Handler(a.listHandler)).Methods("GET")

	// Image routes
	imageRouter := router.PathPrefix("").Subrouter()
	imageRouter.Use(middleware.DeprecatedParams)

	imageRouter.Handle("/{size:[0-9]+}", handler.Handler(a.imageHandler)).Methods("GET")
	imageRouter.Handle("/{width:[0-9]+}/{height:[0-9]+}", handler.Handler(a.imageHandler)).Methods("GET")

	// Image by ID routes
	imageRouter.Handle("/id/{id}/{size:[0-9]+}", handler.Handler(a.imageHandler)).Methods("GET")
	imageRouter.Handle("/id/{id}/{width:[0-9]+}/{height:[0-9]+}", handler.Handler(a.imageHandler)).Methods("GET")

	// Query parameters:
	// ?grayscale - Grayscale the image
	// ?blur - Blur the image
	// ?blur={amount} - Blur the image by {amount}

	// Deprecated query parameters:
	// ?image={id} - Get image by id

	// Deprecated routes
	router.Handle("/list", handler.Handler(a.deprecatedListHandler)).Methods("GET")
	router.Handle("/g/{size:[0-9]+}", handler.Handler(a.deprecatedImageHandler)).Methods("GET")
	router.Handle("/g/{width:[0-9]+}/{height:[0-9]+}", handler.Handler(a.deprecatedImageHandler)).Methods("GET")

	// Set up middleware for adding a request id, handling panics, request logging, and setting CORS headers
	return middleware.AddRequestID(middleware.Recovery(a.Log, middleware.Logger(a.Log, middleware.CORS(router))))
}
