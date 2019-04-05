package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DMarby/picsum-photos/api/handler"
	"github.com/DMarby/picsum-photos/database"
)

const (
	// Default number of items per page
	defaultLimit = 30
	// Max number of items per page
	maxLimit = 100
)

// ListImage contains metadata and download information about an image
type ListImage struct {
	database.Image
	DownloadURL string `json:"download_url"`
}

// Paginated list, with `page` and `limit` query parameters
func (a *API) listHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	limit := getLimit(r)
	page := getPage(r)

	offset := limit * (page - 1)

	databaseList, err := a.Database.List(offset, limit)
	if err != nil {
		a.logError(r, "error getting image list from database", err)
		return handler.InternalServerError()
	}

	list := []ListImage{}

	for _, image := range databaseList {
		list = append(list, ListImage{
			Image: database.Image{
				ID:     image.ID,
				Author: image.Author,
				Width:  image.Width,
				Height: image.Height,
				URL:    image.URL,
			},
			DownloadURL: fmt.Sprintf("%s/id/%s/%d/%d", a.RootURL, image.ID, image.Width, image.Height),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// If we've ran out of items, don't include the next page in the Link header
	end := len(list) < limit
	w.Header().Set("Link", a.getLinkHeader(page, limit, end))

	if err := json.NewEncoder(w).Encode(list); err != nil {
		a.logError(r, "error encoding image list", err)
		return handler.InternalServerError()
	}

	return nil
}

func getLimit(r *http.Request) int {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	return limit
}

func getPage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	return page
}

func (a *API) getLinkHeader(page, limit int, end bool) string {
	// This will return a next even if there's only enough items for a single page, but lets ignore that for now
	if page == 1 {
		return fmt.Sprintf("<%s/v2/list?page=%d&limit=%d>; rel=\"next\"", a.RootURL, page+1, limit)
	}

	if end {
		return fmt.Sprintf("<%s/v2/list?page=%d&limit=%d>; rel=\"prev\"", a.RootURL, page-1, limit)
	}

	return fmt.Sprintf("<%s/v2/list?page=%d&limit=%d>; rel=\"prev\", <%s/v2/list?page=%d&limit=%d>; rel=\"next\"",
		a.RootURL, page-1, limit, a.RootURL, page+1, limit,
	)
}
