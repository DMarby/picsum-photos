package api

import (
	"encoding/json"
	"net/http"

	"github.com/DMarby/picsum-photos/api/handler"
)

func (a *API) healthHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	status := a.HealthChecker.Status()

	if !status.Healthy {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		return handler.InternalServerError()
	}

	return nil
}
