package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RouteMatcher matches routes
type RouteMatcher interface {
	Match(r *http.Request) string
}

// MuxRouteMatcher maxes routes for a mux router
type MuxRouteMatcher struct {
	Router *mux.Router
}

// Match returns the mux route name of a given request
func (m *MuxRouteMatcher) Match(r *http.Request) string {
	var routeMatch mux.RouteMatch
	// The Route can be nil even on a Match, if a NotFoundHandler is specified
	if m.Router.Match(r, &routeMatch) && routeMatch.Route != nil {
		if tmpl, err := routeMatch.Route.GetPathTemplate(); err == nil {
			return tmpl
		}
	}

	return "unknown"
}
