package imageapi

import (
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/params"
	"github.com/gorilla/mux"
)

func (a *API) imageHandler(w http.ResponseWriter, r *http.Request) *handler.Error {

	// Get the path and query parameters
	p, err := params.GetParams(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	// Get the image from the database
	vars := mux.Vars(r)
	imageID := vars["id"]
	databaseImage, handlerErr := a.getImage(r, imageID)
	if handlerErr != nil {
		return handlerErr
	}

	// Validate the parameters
	if err := p.Validate(databaseImage); err != nil {
		return handler.BadRequest(err.Error())
	}

	width, height := p.Dimensions(databaseImage)

	// Build the image task
	task := image.NewTask(databaseImage.ID, width, height, fmt.Sprintf("Picsum ID: %s", databaseImage.ID), getOutputFormat(p.Extension))
	if p.Blur {
		task.Blur(p.BlurAmount)
	}

	if p.Grayscale {
		task.Grayscale()
	}

	// Process the image
	processedImage, err := a.ImageProcessor.ProcessImage(r.Context(), task)
	if err != nil {
		a.logError(r, "error processing image", err)
		return handler.InternalServerError()
	}

	// Set the headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", buildFilename(imageID, p, width, height)))
	w.Header().Set("Content-Type", getContentType(p.Extension))
	w.Header().Set("Cache-Control", "public, max-age=2592000") // Cache for a month
	w.Header().Set("Picsum-ID", databaseImage.ID)

	// Return the image
	w.Write(processedImage)

	return nil
}

func (a *API) getImage(r *http.Request, imageID string) (*database.Image, *handler.Error) {
	databaseImage, err := a.Database.Get(imageID)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, &handler.Error{Message: err.Error(), Code: http.StatusNotFound}
		}

		a.logError(r, "error getting image from database", err)
		return nil, handler.InternalServerError()
	}

	return databaseImage, nil
}

func getOutputFormat(extension string) image.OutputFormat {
	switch extension {
	case ".webp":
		return image.WebP
	default:
		return image.JPEG
	}
}

func getContentType(extension string) string {
	switch extension {
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func buildFilename(imageID string, p *params.Params, width int, height int) string {
	filename := fmt.Sprintf("%s-%dx%d", imageID, width, height)

	if p.Blur {
		filename += fmt.Sprintf("-blur_%d", p.BlurAmount)
	}

	if p.Grayscale {
		filename += "-grayscale"
	}

	filename += p.Extension

	return filename
}
