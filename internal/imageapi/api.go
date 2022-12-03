package imageapi

import (
	"net/http"
	"time"

	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/hmac"
	"github.com/DMarby/picsum-photos/internal/tracing"
	"github.com/rs/cors"

	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/gorilla/mux"
)

// API is a http api
type API struct {
	ImageProcessor image.Processor
	Log            *logger.Logger
	Tracer         *tracing.Tracer
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

	// Image by ID routes
	router.Handle("/id/{id}/{width:[0-9]+}/{height:[0-9]+}{extension:\\..*}", handler.Handler(a.imageHandler)).Methods("GET").Name("imageapi.image")

	// Query parameters:
	// ?grayscale - Grayscale the image
	// ?blur={amount} - Blur the image by {amount}

	// ?hmac - HMAC signature of the path and URL parameters

	// Set up handlers
	cors := cors.New(cors.Options{
		AllowedMethods: []string{"GET"},
		AllowedOrigins: []string{"*"},
		ExposedHeaders: []string{"Content-Type", "Picsum-ID"},
	})

	httpHandler := cors.Handler(router)
	httpHandler = handler.Recovery(a.Log, httpHandler)
	httpHandler = http.TimeoutHandler(httpHandler, a.HandlerTimeout, "Something went wrong. Timed out.")
	httpHandler = handler.Logger(a.Log, httpHandler)

	routeMatcher := &handler.MuxRouteMatcher{Router: router}
	httpHandler = handler.Tracer(a.Tracer, httpHandler, routeMatcher)
	httpHandler = handler.Metrics(httpHandler, routeMatcher)

	return httpHandler
}

// Handle not found errors
var notFoundError = &handler.Error{
	Message: "page not found",
	Code:    http.StatusNotFound,
}

func (a *API) notFoundHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	return notFoundError
}
