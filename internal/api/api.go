package api

import (
	"net/http"
	"path"
	"time"

	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/hmac"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/health"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/gorilla/mux"
)

// API is a http api
type API struct {
	Database        database.Provider
	HealthChecker   *health.Checker
	Log             *logger.Logger
	RootURL         string
	ImageServiceURL string
	StaticPath      string
	HandlerTimeout  time.Duration
	HMAC            *hmac.HMAC
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

	// Image list
	router.Handle("/v2/list", handler.Handler(a.listHandler)).Methods("GET")

	// Query parameters:
	// ?page={page} - What page to display
	// ?limit={limit} - How many entries to display per page

	// Image routes
	oldRouter := router.PathPrefix("").Subrouter()
	oldRouter.Use(a.deprecatedParams)

	oldRouter.Handle("/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.randomImageRedirectHandler)).Methods("GET")
	oldRouter.Handle("/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.randomImageRedirectHandler)).Methods("GET")

	// Image by ID routes
	router.Handle("/id/{id}/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.imageRedirectHandler)).Methods("GET")
	router.Handle("/id/{id}/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.imageRedirectHandler)).Methods("GET")

	// Image info routes
	router.Handle("/id/{id}/info", handler.Handler(a.infoHandler)).Methods("GET")

	// Image by seed routes
	router.Handle("/seed/{seed}/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.seedImageRedirectHandler)).Methods("GET")
	router.Handle("/seed/{seed}/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.seedImageRedirectHandler)).Methods("GET")

	// Query parameters:
	// ?grayscale - Grayscale the image
	// ?blur - Blur the image
	// ?blur={amount} - Blur the image by {amount}

	// Deprecated query parameters:
	// ?image={id} - Get image by id

	// Deprecated routes
	router.Handle("/list", handler.Handler(a.deprecatedListHandler)).Methods("GET")
	router.Handle("/g/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.deprecatedImageHandler)).Methods("GET")
	router.Handle("/g/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.deprecatedImageHandler)).Methods("GET")

	// Static files
	router.HandleFunc("/", serveFile(path.Join(a.StaticPath, "index.html")))
	router.HandleFunc("/images", serveFile(path.Join(a.StaticPath, "images.html")))
	router.HandleFunc("/favicon.ico", serveFile(path.Join(a.StaticPath, "assets/images/favicon/favicon.ico")))
	router.PathPrefix("/assets/").HandlerFunc(fileHeaders(http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(a.StaticPath, "assets/")))).ServeHTTP))

	// Set up handlers for adding a request id, handling panics, request logging, setting CORS headers, and handler execution timeout
	return handler.AddRequestID(handler.Recovery(a.Log, handler.Logger(a.Log, handler.CORS(nil, http.TimeoutHandler(router, a.HandlerTimeout, "Something went wrong. Timed out.")))))
}

// Handle not found errors
var notFoundError = &handler.Error{
	Message: "page not found",
	Code:    http.StatusNotFound,
}

func (a *API) notFoundHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	return notFoundError
}

// Set headers for static file handlers
func fileHeaders(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		handler(w, r)
	}
}

// Serve a static file
func serveFile(name string) func(w http.ResponseWriter, r *http.Request) {
	return fileHeaders(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}
