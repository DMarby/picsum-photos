package api

import (
	"net/http"

	"github.com/DMarby/picsum-photos/queue"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// API is a http api
type API struct {
	workerQueue *queue.Queue
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	images := make([]string, 0)
	render.JSON(w, r, images)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	size := chi.URLParam(r, "size") // TODO: check if "" or cast?
	w.Write([]byte(size))
}

// New instantiates and returns a new api
func New(workerQueue *queue.Queue) *API {
	return &API{
		workerQueue: workerQueue,
	}
}

// Router returns a http router
func (api *API) Router() http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger) // TODO: Use logrus?
	router.Use(middleware.Recoverer)

	// TODO: Healthcheck for LBs/autoscaling

	// Image list
	router.Get("/list", listHandler)

	// Image routes
	router.Get("/{size}", imageHandler)
	router.Get("/{width}/{height}", imageHandler)

	// Deprecated routes
	router.Get("/g/{size}", imageHandler)
	router.Get("/g/{width}/{height}", imageHandler)

	// TODO: Ensure we set CORS correctly
	// TODO: Either serve static pages for index/images, or let nginx take care of it
	// TODO: Gzip everything, or let nginx handle that too
	// TODO: Custom 404 handler for everything else?

	return router
}
