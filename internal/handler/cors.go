package handler

import (
	"net/http"
	"strings"
)

// CORS is a handler for setting CORS headers
// Based on https://github.com/gorilla/handlers/blob/master/cors.go
func CORS(exposedHeaders []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == "OPTIONS" {
			if _, ok := r.Header["Access-Control-Request-Method"]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			method := r.Header.Get("Access-Control-Request-Method")
			if method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			if headers := r.Header.Get("Access-Control-Request-Headers"); headers != "" {
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET")
		} else {
			// Expose headers
			if exposedHeaders != nil {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(exposedHeaders, ", "))
			}

			next.ServeHTTP(w, r)
		}
	})
}
