package api

import (
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/api/handler"
	"github.com/DMarby/picsum-photos/api/params"
	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/image"
	"github.com/gorilla/mux"
)

func (a *API) imageHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	width, height, err := params.GetSize(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	grayscale, blur, blurAmount := params.GetQueryParams(r)
	vars := mux.Vars(r)
	imageID, ok := vars["id"]

	if !ok || imageID == "" { // Redirect to a random image when no image is specified
		randomImage, err := a.Database.GetRandom()
		if err != nil {
			a.logError(r, "error getting random image from database", err)
			return handler.InternalServerError()
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.Redirect(w, r, fmt.Sprintf("/id/%s/%d/%d%s", randomImage, width, height, params.BuildQuery(grayscale, blur, blurAmount)), http.StatusFound)
		return nil
	}

	databaseImage, err := a.Database.Get(imageID)
	if err != nil {
		if err == database.ErrNotFound {
			return &handler.Error{Message: err.Error(), Code: http.StatusNotFound}
		}

		a.logError(r, "error getting image from database", err)
		return handler.InternalServerError()
	}

	if err := params.ValidateParams(a.MaxImageSize, databaseImage, width, height, blur, blurAmount); err != nil {
		return handler.BadRequest(err.Error())
	}

	// Default to the image width/height if 0 is passed
	if width == 0 {
		width = databaseImage.Width
	}

	if height == 0 {
		height = databaseImage.Height
	}

	imageBuffer, err := a.Cache.Get(imageID)
	if err != nil {
		a.logError(r, "error getting image from cache", err)
		return handler.InternalServerError()
	}

	task := image.NewTask(imageBuffer, width, height)

	if grayscale {
		task.Grayscale()
	}

	if blur {
		task.Blur(blurAmount)
	}

	processedImage, err := a.ImageProcessor.ProcessImage(task)
	if err != nil {
		a.logError(r, "error processing image", err)
		return handler.InternalServerError()
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for a day
	w.Write(processedImage)

	return nil
}
