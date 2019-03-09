package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/api/handler"
	"github.com/DMarby/picsum-photos/database"
)

// ListImage contains metadata and download information about an image
type ListImage struct {
	database.Image
	DownloadURL string `json:"download_url"`
}

func (a *API) listHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	databaseList, err := a.Database.List()
	if err != nil {
		a.logError(r, "error getting image list from database", err)
		return handler.InternalServerError()
	}

	var list []ListImage

	for _, image := range databaseList {
		list = append(list, ListImage{
			Image: database.Image{
				ID:     image.ID,
				Author: image.Author,
				Width:  image.Width,
				Height: image.Height,
				URL:    image.URL,
			},
			DownloadURL: fmt.Sprintf("https://picsum.photos/id/%s/%d/%d", image.ID, image.Width, image.Height),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		a.logError(r, "error encoding image list", err)
		return handler.InternalServerError()
	}

	return nil
}
