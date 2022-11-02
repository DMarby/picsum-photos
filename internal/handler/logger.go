package handler

import (
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/felixge/httpsnoop"
)

// Logger is a handler that logs requests using Zap
func Logger(log *logger.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := GetReqID(ctx)

		respMetrics := httpsnoop.CaptureMetricsFn(w, func(ww http.ResponseWriter) {
			h.ServeHTTP(ww, r)
		})

		logFields := []interface{}{
			"request-id", id,
			"http-method", r.Method,
			"remote-addr", r.RemoteAddr,
			"user-agent", r.UserAgent(),
			"uri", r.URL.String(),
			"status-code", respMetrics.Code,
			"elapsed", fmt.Sprintf("%.9fs", respMetrics.Duration.Seconds()),
		}

		switch {
		case respMetrics.Code >= 500:
			log.Errorw("Request completed", logFields...)
		default:
			log.Debugw("Request completed", logFields...)
		}
	})
}

// LogFields logs the given keys and values for a request
func LogFields(r *http.Request, keysAndValues ...interface{}) []interface{} {
	ctx := r.Context()
	id := GetReqID(ctx)

	return append([]interface{}{"request-id", id}, keysAndValues...)
}
