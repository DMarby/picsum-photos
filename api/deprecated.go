package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DMarby/picsum-photos/api/handler"
	"github.com/DMarby/picsum-photos/api/params"
)

// DeprecatedImage contains info about an image, in the old deprecated /list style
type DeprecatedImage struct {
	Format    string `json:"format"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Filename  string `json:"filename"`
	ID        int    `json:"id"`
	Author    string `json:"author"`
	AuthorURL string `json:"author_url"`
	PostURL   string `json:"post_url"`
}

func (a *API) deprecatedListHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	list, err := a.Database.List()
	if err != nil {
		a.logError(r, "error getting image list from database", err)
		return handler.InternalServerError()
	}

	var images []DeprecatedImage
	for _, image := range list {
		numericID, err := strconv.Atoi(image.ID)
		if err != nil {
			continue
		}

		images = append(images, DeprecatedImage{
			Format:    "jpeg",
			Width:     image.Width,
			Height:    image.Height,
			Filename:  fmt.Sprintf("%s.jpeg", image.ID),
			ID:        numericID,
			Author:    image.Author,
			AuthorURL: image.URL,
			PostURL:   image.URL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(images); err != nil {
		a.logError(r, "error encoding deprecate image list", err)
		return handler.InternalServerError()
	}

	return nil
}

// Handles deprecated image routes
func (a *API) deprecatedImageHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	width, height, err := params.GetSize(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	_, blur, blurAmount := params.GetQueryParams(r)

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Look for the deprecated ?image query parameter
	if id := r.URL.Query().Get("image"); id != "" {
		http.Redirect(w, r, fmt.Sprintf("/id/%s/%d/%d%s", id, width, height, params.BuildQuery(true, blur, blurAmount)), http.StatusFound)
		return nil
	}

	http.Redirect(w, r, fmt.Sprintf("/%d/%d%s", width, height, params.BuildQuery(true, blur, blurAmount)), http.StatusFound)
	return nil
}
