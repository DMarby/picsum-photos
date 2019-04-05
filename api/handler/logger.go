package handler

import (
	"net/http"
	"time"

	"github.com/DMarby/picsum-photos/logger"
)

// Logger is a handler that logs requests using Zap
func Logger(log *logger.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := GetReqID(ctx)

		fields := []interface{}{
			"request-id", id,
			"http-method", r.Method,
			"remote-addr", r.RemoteAddr,
			"user-agent", r.UserAgent(),
			"uri", r.RequestURI,
		}

		log.Debugw("request started", fields...)

		start := time.Now()
		h.ServeHTTP(w, r)

		log.Debugw("request completed", append(fields, "elapsed-ms", float64(time.Since(start).Nanoseconds())/1000000.0)...)
	})
}
