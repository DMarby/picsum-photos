package handler

import "net/http"

// CORS is a handler for setting CORS headers
// Based on https://github.com/gorilla/handlers/blob/master/cors.go
func CORS(next http.Handler) http.Handler {
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
			// Expose the Link header used for pagination
			// And the Picsum-ID used to get the ID for an image
			w.Header().Set("Access-Control-Expose-Headers", "Link, Picsum-ID")

			next.ServeHTTP(w, r)
		}
	})
}
