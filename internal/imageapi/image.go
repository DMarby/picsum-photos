package imageapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/params"
	"github.com/gorilla/mux"
)

func (a *API) imageHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	// Validate the path and query parameters
	valid, err := params.ValidateHMAC(a.HMAC, r)
	if err != nil {
		return handler.InternalServerError()
	}

	if !valid {
		return handler.BadRequest("Invalid parameters")
	}

	// Get the path and query parameters
	p, err := params.GetParams(r)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	// Get the image ID from the path param
	vars := mux.Vars(r)
	imageID := vars["id"]

	// Build the image task
	task := image.NewTask(imageID, p.Width, p.Height, fmt.Sprintf("Picsum ID: %s", imageID), getOutputFormat(p.Extension))
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
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", buildFilename(imageID, p)))
	w.Header().Set("Content-Type", getContentType(p.Extension))
	w.Header().Set("Content-Length", strconv.Itoa(len(processedImage)))
	w.Header().Set("Cache-Control", "public, max-age=2592000") // Cache for a month
	w.Header().Set("Picsum-ID", imageID)

	// Return the image
	w.Write(processedImage)

	return nil
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

func buildFilename(imageID string, p *params.Params) string {
	filename := fmt.Sprintf("%s-%dx%d", imageID, p.Width, p.Height)

	if p.Blur {
		filename += fmt.Sprintf("-blur_%d", p.BlurAmount)
	}

	if p.Grayscale {
		filename += "-grayscale"
	}

	filename += p.Extension

	return filename
}
