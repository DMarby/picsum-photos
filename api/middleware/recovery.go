package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/DMarby/picsum-photos/logger"
)

// Recovery is a middleware for handling panics
func Recovery(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				ctx := r.Context()
				id := GetReqID(ctx)
				log.Errorw("panic handling request",
					"request-id", id,
					"stacktrace", string(debug.Stack()),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
