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

		responseWriter := &loggingResponseWriter{w, -1}
		start := time.Now()
		h.ServeHTTP(responseWriter, r)

		log.Debugw("request completed", append(fields, "status-code", responseWriter.statusCode, "elapsed-ms", float64(time.Since(start).Nanoseconds())/1000000.0)...)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}
