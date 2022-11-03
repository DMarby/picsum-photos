package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/DMarby/picsum-photos/internal/database"

	_ "image/jpeg"
)

// Comandline flags
var (
	imagePath         = flag.String("image-path", ".", "path to image directory")
	imageManifestPath = flag.String("image-manifest-path", "./image-manifest.json", "path to the image manifest to update")
)

func main() {
	flag.Parse()

	resolvedManifestPath, err := filepath.Abs(*imageManifestPath)
	if err != nil {
		log.Fatal(err)
	}

	manifestData, err := os.ReadFile(resolvedManifestPath)
	if err != nil {
		log.Fatal(err)
	}

	var images []database.Image
	err = json.Unmarshal(manifestData, &images)
	if err != nil {
		log.Fatal(err)
	}

	for i, img := range images {
		resolvedImagePath, err := filepath.Abs(filepath.Join(*imagePath, fmt.Sprintf("%s.jpg", img.ID)))
		if err != nil {
			log.Fatal(err)
		}

		reader, err := os.Open(resolvedImagePath)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		imageMetadata, _, err := image.DecodeConfig(reader)
		if err != nil {
			log.Fatal(err)
		}

		images[i].Width = imageMetadata.Width
		images[i].Height = imageMetadata.Height
	}

	file, _ := os.OpenFile(resolvedManifestPath, os.O_WRONLY, 0644)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(images); err != nil {
		log.Fatal(err)
	}
}
