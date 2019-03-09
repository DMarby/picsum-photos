package handler

import (
	"encoding/json"
	"net/http"
)

// Error is the message and http status code to return
type Error struct {
	Message string
	Code    int
}

// InternalServerError is a convenience function for returning an internal server error
func InternalServerError() *Error {
	return &Error{
		Message: "Something went wrong",
		Code:    http.StatusInternalServerError,
	}
}

// BadRequest is a convenience function for returning a bad request error
func BadRequest(message string) *Error {
	return &Error{
		Message: message,
		Code:    http.StatusBadRequest,
	}
}

const jsonMediaType = "application/json"

// Handler wraps a http handler and deals with responding to errors
type Handler func(w http.ResponseWriter, r *http.Request) *Error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		if r.Header.Get("accept") == jsonMediaType {
			var data = struct {
				Error string `json:"error"`
			}{err.Message}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(data); err != nil {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Message, err.Code)
		}
	}
}
