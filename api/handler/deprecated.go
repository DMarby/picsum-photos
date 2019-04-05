package handler

import (
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/api/params"
)

// DeprecatedParams is a handler to handle deprecated query params for regular routes
func DeprecatedParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Look for the deprecated ?image query parameter
		if id := r.URL.Query().Get("image"); id != "" {
			width, height, err := params.GetSize(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			grayscale, blur, blurAmount := params.GetQueryParams(r)

			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			http.Redirect(w, r, fmt.Sprintf("/id/%s/%d/%d%s", id, width, height, params.BuildQuery(grayscale, blur, blurAmount)), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
