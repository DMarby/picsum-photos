package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/oceanicdev/chi-param" // TODO: Get rid of chi-param, just parse it ourselves
)

// API is a http api
type API struct {
	imageProcessor image.Processor
	storage        storage.Provider
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "picsum context key " + k.name
}

// TODO: Look for projects on github using the /list endpoint to see if any of the now removed/renamed fields are used, to consider whether we should still support them
// TODO: Maybe just on the unsplash.it domain?
func (a *API) listHandler(w http.ResponseWriter, r *http.Request) {
	list, err := a.storage.List()
	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}

	render.JSON(w, r, list)
}

func imageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Look for the id param
		if id, err := param.String(r, "id"); err == nil && id != "" {
			r = r.WithContext(context.WithValue(ctx, contextKey{"imageID"}, id))
		}

		next.ServeHTTP(w, r)
	})
}

func getImageID(ctx context.Context) string {
	if imageID, ok := ctx.Value(contextKey{"imageID"}).(string); ok {
		return imageID
	}

	return ""
}

var errInvalidSize = errors.New("Invalid size")

// TODO: Cap size to 5000 or native size like in the last API too?
func sizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var width int
		var height int

		// Check for the size parameter first
		if size, err := param.Int(r, "size"); err == nil {
			width, height = size, size
		} else {
			var err error
			// If size doesn't exist, check for width/height
			width, err = param.Int(r, "width")
			if err != nil {
				http.Error(w, errInvalidSize.Error(), 400)
				return
			}

			height, err = param.Int(r, "height")
			if err != nil {
				http.Error(w, errInvalidSize.Error(), 400)
				return
			}
		}

		if width < 1 || height < 1 {
			http.Error(w, errInvalidSize.Error(), 400)
			return
		}

		// TODO: Simplify?
		ctx = context.WithValue(ctx, contextKey{"width"}, width)
		ctx = context.WithValue(ctx, contextKey{"height"}, height)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
func getSize(ctx context.Context) (width int, height int) {
	// TODO: error handling needed?
	ctxWidth, ok := ctx.Value(contextKey{"width"}).(int)
	if ok {
		width = ctxWidth
	}

	ctxHeight, ok := ctx.Value(contextKey{"height"}).(int)
	if ok {
		height = ctxHeight
	}

	return width, height
}

func queryParamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var grayscale bool
		var blur bool

		if _, ok := r.URL.Query()["grayscale"]; ok {
			grayscale = true
		}

		if _, ok := r.URL.Query()["blur"]; ok {
			blur = true
		}

		// TODO: Simplify?
		ctx = context.WithValue(ctx, contextKey{"grayscale"}, grayscale)
		ctx = context.WithValue(ctx, contextKey{"blur"}, blur)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func getQueryParams(ctx context.Context) (grayscale bool, blur bool) {
	// TODO: error handling needed?
	ctxGrayscale, ok := ctx.Value(contextKey{"grayscale"}).(bool)
	if ok {
		grayscale = ctxGrayscale
	}

	ctxBlur, ok := ctx.Value(contextKey{"blur"}).(bool)
	if ok {
		blur = ctxBlur
	}

	return grayscale, blur
}

func addParam(buf *bytes.Buffer, param string) {
	if buf.Len() > 0 {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}

	buf.WriteString(param)
}

func buildQuery(grayscale bool, blur bool) string {
	if !grayscale && !blur {
		return ""
	}

	var buf bytes.Buffer

	if grayscale {
		addParam(&buf, "grayscale")
	}

	if blur {
		addParam(&buf, "blur")
	}

	return buf.String()
}

func (a *API) imageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// TODO: Simplify to one context reader?
	width, height := getSize(ctx)
	grayscale, blur := getQueryParams(ctx)
	imageID := getImageID(ctx)

	if imageID == "" { // Redirect to a random image when no image is specified
		randomImage, err := a.storage.GetRandom()
		if err != nil {
			log.Print(err) // TODO: Logrus
			http.Error(w, "Something went wrong", 500)
			return
		}

		// TODO: Set cache headers so this doesn't get cached, make sure resulting url is
		http.Redirect(w, r, fmt.Sprintf("/id/%s/%d/%d%s", randomImage, width, height, buildQuery(grayscale, blur)), http.StatusFound)
		return
	}

	imageBuffer, err := a.storage.Get(imageID)
	if err != nil {
		if err == storage.ErrNotFound { // TODO: Log all response codes
			http.Error(w, err.Error(), 404)
		} else {
			log.Print(err) // TODO: Logrus
			http.Error(w, "Something went wrong", 500)
		}

		return
	}

	// TODO: Should this be a builder?
	task := image.NewTask(imageBuffer, width, height)

	if grayscale {
		task.Grayscale() // TODO: Should we need to reassign here instead? Or not be able to chain? Is it weird that it's both mutable and chainable?
	}

	if blur {
		task.Blur(5) // TODO: How much blur?
	}

	processedImage, err := a.imageProcessor.ProcessImage(task)
	if err != nil {
		log.Print(err) // TODO: Logrus
		http.Error(w, "Something went wrong", 500)
		return
	}

	// TODO: Cache-control headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(processedImage)
}

// Middleware to handle deprecated query params
func deprecationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Look for the deprecated ?image query parameter
		if id, err := param.QueryString(r, "image"); err == nil && id != "" {
			ctx := r.Context()
			// TODO: Simplify to one context reader?
			width, height := getSize(ctx)
			grayscale, blur := getQueryParams(ctx)

			http.Redirect(w, r, fmt.Sprintf("/id/%s/%d/%d%s", id, width, height, buildQuery(grayscale, blur)), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handlers deprecated routes
func (a *API) deprecationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// TODO: Simplify to one context reader?
	width, height := getSize(ctx)
	_, blur := getQueryParams(ctx)

	if id, err := param.QueryString(r, "image"); err == nil && id != "" {
		// TODO: Set cache headers so this doesn't get cached, make sure resulting url is
		http.Redirect(w, r, fmt.Sprintf("/id/%s/%d/%d%s", id, width, height, buildQuery(true, blur)), http.StatusFound)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/%d/%d%s", width, height, buildQuery(true, blur)), http.StatusFound)
	}
}

// New instantiates and returns a new api
func New(imageProcessor image.Processor, storage storage.Provider) *API {
	return &API{
		imageProcessor: imageProcessor,
		storage:        storage,
	}
}

// Router returns a http router
func (a *API) Router() http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger) // TODO: Use logrus or https://github.com/pressly/lg or something
	router.Use(middleware.Recoverer)
	// router.Use(middleware.RedirectSlashes) // TODO: Needed? Or just StripSlashes? Add tests for trailing / and so on, or maybe ignore?
	// TODO: Timeout?

	// TODO: Healthcheck for LBs/autoscaling

	// Image list
	router.Get("/list", a.listHandler)

	// Image routes
	router.Group(func(router chi.Router) {
		router.Use(sizeMiddleware)
		router.Use(queryParamMiddleware)
		router.Use(imageMiddleware)

		router.Group(func(router chi.Router) {
			router.Use(deprecationMiddleware)

			router.Get("/{size:[0-9]+}", a.imageHandler)
			router.Get("/{width:[0-9]+}/{height:[0-9]+}", a.imageHandler)

			// Image by ID routes
			router.Get("/id/{id}/{size:[0-9]+}", a.imageHandler)
			router.Get("/id/{id}/{width:[0-9]+}/{height:[0-9]+}", a.imageHandler)

			// TODO: Image by Hash routes
			// router.Get("/hash/{hash}/{size:[0-9]+}", a.imageHandler)
			// router.Get("/hash/{hash}/{width:[0-9]+}/{height:[0-9]+}", a.imageHandler)

			// Query parameters:
			// TODO: ?text=size - Overlay {width}x{height} on top of image
			// TODO: ?text={text} - Overlay text on top of image
			// ?grayscale - Grayscale the image
			// ?blur - Blur the image

			// Deprecated query parameters:
			// ?image={id} - Get image by id
		})

		// Deprecated routes
		router.Get("/g/{size:[0-9]+}", a.deprecationHandler)
		router.Get("/g/{width:[0-9]+}/{height:[0-9]+}", a.deprecationHandler)
	})

	// TODO: Ensure we set CORS correctly
	// TODO: Either serve static pages for index/images, or let nginx take care of it
	// TODO: Gzip everything, or let nginx handle that too
	// TODO: Custom 404 handler for everything else + tests NotFound func
	// TODO: Graceful shutdown
	// TODO: Make sure requesting dimensions larger than the original image works as well
	// TODO: Add tests for headers (cors, cache control, etc)

	return router
}
