package api

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/hmac"
	"github.com/DMarby/picsum-photos/internal/tracing"
	"github.com/rs/cors"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/gorilla/mux"

	_ "embed"
)

//go:embed web/embed
var static embed.FS

// API is a http api
type API struct {
	Database        database.Provider
	Log             *logger.Logger
	Tracer          *tracing.Tracer
	RootURL         string
	ImageServiceURL string
	HandlerTimeout  time.Duration
	HMAC            *hmac.HMAC
}

// Utility methods for logging
func (a *API) logError(r *http.Request, message string, err error) {
	a.Log.Errorw(message, handler.LogFields(r, "error", err)...)
}

// Router returns a http router
func (a *API) Router() (http.Handler, error) {
	router := mux.NewRouter()

	router.NotFoundHandler = handler.Handler(a.notFoundHandler)

	// Redirect trailing slashes
	router.StrictSlash(true)

	// Image list
	router.Handle("/v2/list", handler.Handler(a.listHandler)).Methods("GET").Name("api.list")

	// Query parameters:
	// ?page={page} - What page to display
	// ?limit={limit} - How many entries to display per page

	// Image routes
	oldRouter := router.PathPrefix("").Subrouter()
	oldRouter.Use(a.deprecatedParams)

	oldRouter.Handle("/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.randomImageRedirectHandler)).Methods("GET").Name("api.randomImageRedirect")
	oldRouter.Handle("/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.randomImageRedirectHandler)).Methods("GET").Name("api.randomImageRedirect")

	// Image by ID routes
	router.Handle("/id/{id}/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.imageRedirectHandler)).Methods("GET").Name("api.imageRedirect")
	router.Handle("/id/{id}/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.imageRedirectHandler)).Methods("GET").Name("api.imageRedirect")

	// Image info routes
	router.Handle("/id/{id}/info", handler.Handler(a.infoHandler)).Methods("GET").Name("api.info")
	router.Handle("/seed/{seed}/info", handler.Handler(a.infoSeedHandler)).Methods("GET").Name("api.infoSeed")

	// Image by seed routes
	router.Handle("/seed/{seed}/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.seedImageRedirectHandler)).Methods("GET").Name("api.seedImageRedirect")
	router.Handle("/seed/{seed}/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.seedImageRedirectHandler)).Methods("GET").Name("api.seedImageRedirect")

	// Query parameters:
	// ?grayscale - Grayscale the image
	// ?blur - Blur the image
	// ?blur={amount} - Blur the image by {amount}

	// Deprecated query parameters:
	// ?image={id} - Get image by id

	// Deprecated routes
	router.Handle("/list", handler.Handler(a.deprecatedListHandler)).Methods("GET").Name("api.deprecatedList")
	router.Handle("/g/{size:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.deprecatedImageHandler)).Methods("GET").Name("api.deprecatedImage")
	router.Handle("/g/{width:[0-9]+}/{height:[0-9]+}{extension:(?:\\..*)?}", handler.Handler(a.deprecatedImageHandler)).Methods("GET").Name("api.deprecatedImage")

	// Static files
	staticFS, err := fs.Sub(static, "web/embed")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(staticFS))

	router.HandleFunc("/", serveFile(fileServer, "/")).Name("api.serveFile")
	router.HandleFunc("/images", serveFile(fileServer, "/images.html")).Name("api.serveFile")
	router.HandleFunc("/favicon.ico", serveFile(fileServer, "/favicon.ico")).Name("api.serveFile")
	router.PathPrefix("/assets/").HandlerFunc(fileHeaders(fileServer.ServeHTTP)).Name("api.serveFile")

	// Set up handlers
	cors := cors.New(cors.Options{
		AllowedMethods: []string{"GET"},
		AllowedOrigins: []string{"*"},
	})

	httpHandler := cors.Handler(router)
	httpHandler = handler.Recovery(a.Log, httpHandler)
	httpHandler = http.TimeoutHandler(httpHandler, a.HandlerTimeout, "Something went wrong. Timed out.")
	httpHandler = handler.Logger(a.Log, httpHandler)

	routeMatcher := &handler.MuxRouteMatcher{Router: router}
	httpHandler = handler.Tracer(a.Tracer, httpHandler, routeMatcher)
	httpHandler = handler.Metrics(httpHandler, routeMatcher)

	return httpHandler, nil
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
		w.Header().Set("Cache-Control", "public, max-age=7200, stale-while-revalidate=60, stale-if-error=43200")
		handler(w, r)
	}
}

// Serve a static file
func serveFile(h http.Handler, name string) func(w http.ResponseWriter, r *http.Request) {
	return fileHeaders(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = name
		h.ServeHTTP(w, r)
	})
}
