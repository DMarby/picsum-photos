package handler

import (
	"encoding/json"
	"net/http"

	"github.com/DMarby/picsum-photos/internal/health"
)

// Health is a handler for health check status
func Health(healthChecker *health.Checker) Handler {
	return Handler(newHandler(healthChecker))
}

func newHandler(healthChecker *health.Checker) func(w http.ResponseWriter, r *http.Request) *Error {
	return func(w http.ResponseWriter, r *http.Request) *Error {
		status := healthChecker.Status()

		if !status.Healthy {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			return InternalServerError()
		}

		return nil
	}
}
