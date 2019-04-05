package handler

import "net/http"

// CORS is a handler for setting CORS headers
// Based on https://github.com/gorilla/handlers/blob/master/cors.go
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == "" {
			next.ServeHTTP(w, r)
			return
		}

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
			next.ServeHTTP(w, r)
		}
	})
}
