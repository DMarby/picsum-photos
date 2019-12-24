package handler

import (
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/internal/params"
)

// DeprecatedParams is a handler to handle deprecated query params for regular routes
func DeprecatedParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Look for the deprecated ?image query parameter
		if id := r.URL.Query().Get("image"); id != "" {
			p, err := params.GetParams(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header()["Content-Type"] = nil
			http.Redirect(w, r, fmt.Sprintf("/id/%s/%d/%d%s%s", id, p.Width, p.Height, p.Extension, params.BuildQuery(p.Grayscale, p.Blur, p.BlurAmount)), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
