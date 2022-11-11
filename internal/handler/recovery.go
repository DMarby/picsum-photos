package handler

import (
	"net/http"
	"runtime/debug"

	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/tracing"
)

// Recovery is a handler for handling panics
func Recovery(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				ctx := r.Context()
				traceID, spanID := tracing.TraceInfo(ctx)
				log.Errorw("panic handling request",
					"trace-id", traceID,
					"span-id", spanID,
					"stacktrace", string(debug.Stack()),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
