package params

import (
	"net/http"
	"net/url"

	"github.com/DMarby/picsum-photos/internal/hmac"
)

// HMAC generates and appends an HMAC to a URL path + query params
func HMAC(h *hmac.HMAC, path string, query url.Values) (string, error) {
	hmac, err := h.Create(path + BuildQuery(query))
	if err != nil {
		return "", err
	}

	query.Set("hmac", hmac)
	return path + BuildQuery(query), nil
}

// ValidateHMAC validates the URL path/query params, given an hmac in a query parameter named hmac
func ValidateHMAC(h *hmac.HMAC, r *http.Request) (bool, error) {
	// Get the query params in the request
	query := r.URL.Query()

	// Get the HMAC query param and remove it from the request query params
	hmac := query.Get("hmac")
	query.Del("hmac")

	encodedQuery := BuildQuery(query)
	return h.Validate(r.URL.Path+encodedQuery, hmac)
}
