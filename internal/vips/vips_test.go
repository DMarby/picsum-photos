package vips_test

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/vips"
	"go.uber.org/zap"

	"testing"
)

func TestVips(t *testing.T) {
	imageBuffer := setup(t)
	defer vips.Shutdown()

	t.Run("SaveToJpegBuffer", func(t *testing.T) {
		t.Run("saves an image to buffer", func(t *testing.T) {
			_, err := vips.SaveToJpegBuffer(resizeImage(t, imageBuffer))
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("errors on an invalid image", func(t *testing.T) {
			_, err := vips.SaveToJpegBuffer(vips.NewEmptyImage())
			if err == nil || !strings.Contains(err.Error(), "error saving to jpeg buffer") || !strings.Contains(err.Error(), "vips_image_pio_input: no image data") {
				t.Error(err)
			}
		})
	})

	t.Run("SaveToWebPBuffer", func(t *testing.T) {
		t.Run("saves an image to buffer", func(t *testing.T) {
			_, err := vips.SaveToWebPBuffer(resizeImage(t, imageBuffer))
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("errors on an invalid image", func(t *testing.T) {
			_, err := vips.SaveToWebPBuffer(vips.NewEmptyImage())
			if err == nil || !strings.Contains(err.Error(), "error saving to webp buffer") || !strings.Contains(err.Error(), "vips_image_pio_input: no image data") {
				t.Error(err)
			}
		})
	})

	t.Run("ResizeImage", func(t *testing.T) {
		t.Run("loads and resizes an image as jpeg", func(t *testing.T) {
			image, err := vips.ResizeImage(imageBuffer, 500, 500)
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToJpegBuffer(image)
			resultFixture := readFixture("resize", "jpg")
			if !reflect.DeepEqual(buf, resultFixture) {
				t.Error("image data doesn't match")
			}
		})

		t.Run("loads and resizes an image as webp", func(t *testing.T) {
			image, err := vips.ResizeImage(imageBuffer, 500, 500)
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToWebPBuffer(image)
			resultFixture := readFixture("resize", "webp")
			if !reflect.DeepEqual(buf, resultFixture) {
				t.Error("image data doesn't match")
			}
		})

		t.Run("errors when given an empty buffer", func(t *testing.T) {
			var buf []byte
			_, err := vips.ResizeImage(buf, 500, 500)
			if err == nil || err.Error() != "empty buffer" {
				t.Error(err)
			}
		})

		t.Run("errors when given an invalid image", func(t *testing.T) {
			_, err := vips.ResizeImage(make([]byte, 5), 500, 500)
			if err == nil || err.Error() != "error processing image from buffer VipsForeignLoad: buffer is not in a known format\n" {
				t.Error(err)
			}
		})
	})

	t.Run("Grayscale", func(t *testing.T) {
		t.Run("converts an image to grayscale as jpeg", func(t *testing.T) {
			image, err := vips.Grayscale(resizeImage(t, imageBuffer))
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToJpegBuffer(image)
			resultFixture := readFixture("grayscale", "jpg")
			if !reflect.DeepEqual(buf, resultFixture) {
				t.Error("image data doesn't match")
			}
		})

		t.Run("converts an image to grayscale as webp", func(t *testing.T) {
			image, err := vips.Grayscale(resizeImage(t, imageBuffer))
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToWebPBuffer(image)
			resultFixture := readFixture("grayscale", "webp")
			if !reflect.DeepEqual(buf, resultFixture) {
				t.Error("image data doesn't match")
			}
		})

		t.Run("errors when given an invalid image", func(t *testing.T) {
			_, err := vips.Grayscale(vips.NewEmptyImage())
			if err == nil || err.Error() != "error changing image colorspace vips_image_pio_input: no image data\n" {
				t.Error(err)
			}
		})
	})

	t.Run("Blur", func(t *testing.T) {
		t.Run("blurs an image as jpeg", func(t *testing.T) {
			image, err := vips.Blur(resizeImage(t, imageBuffer), 5)
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToJpegBuffer(image)
			resultFixture := readFixture("blur", "jpg")
			if !reflect.DeepEqual(buf, resultFixture) {
				t.Error("image data doesn't match")
			}
		})

		t.Run("blurs an image as webp", func(t *testing.T) {
			image, err := vips.Blur(resizeImage(t, imageBuffer), 5)
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToWebPBuffer(image)
			resultFixture := readFixture("blur", "webp")
			if !reflect.DeepEqual(buf, resultFixture) {
				t.Error("image data doesn't match")
			}
		})

		t.Run("errors when given an invalid image", func(t *testing.T) {
			_, err := vips.Blur(vips.NewEmptyImage(), 5)
			if err == nil || err.Error() != "error applying blur to image vips_image_pio_input: no image data\n" {
				t.Error(err)
			}
		})
	})
}

// Utility function for regenerating the fixtures
func TestFixtures(t *testing.T) {
	if os.Getenv("GENERATE_FIXTURES") != "1" {
		t.SkipNow()
	}

	imageBuffer := setup(t)
	defer vips.Shutdown()

	// Resize
	image, _ := vips.ResizeImage(imageBuffer, 500, 500)
	resizeJpeg, _ := vips.SaveToJpegBuffer(image)
	os.WriteFile(fixturePath("resize", "jpg"), resizeJpeg, 0644)

	image, _ = vips.ResizeImage(imageBuffer, 500, 500)
	resizeWebP, _ := vips.SaveToWebPBuffer(image)
	os.WriteFile(fixturePath("resize", "webp"), resizeWebP, 0644)

	// Grayscale
	image, _ = vips.Grayscale(resizeImage(t, imageBuffer))
	grayscaleJpeg, _ := vips.SaveToJpegBuffer(image)
	os.WriteFile(fixturePath("grayscale", "jpg"), grayscaleJpeg, 0644)

	image, _ = vips.Grayscale(resizeImage(t, imageBuffer))
	grayscaleWebP, _ := vips.SaveToWebPBuffer(image)
	os.WriteFile(fixturePath("grayscale", "webp"), grayscaleWebP, 0644)

	// Blur
	image, _ = vips.Blur(resizeImage(t, imageBuffer), 5)
	blurJpeg, _ := vips.SaveToJpegBuffer(image)
	os.WriteFile(fixturePath("blur", "jpg"), blurJpeg, 0644)

	image, _ = vips.Blur(resizeImage(t, imageBuffer), 5)
	blurWebP, _ := vips.SaveToWebPBuffer(image)
	os.WriteFile(fixturePath("blur", "webp"), blurWebP, 0644)
}

func setup(t *testing.T) []byte {
	log := logger.New(zap.FatalLevel)
	defer log.Sync()

	err := vips.Initialize(log)
	if err != nil {
		t.Fatal(err)
	}

	imageBuffer, err := os.ReadFile("../../test/fixtures/fixture.jpg")
	if err != nil {
		t.Fatal(err)
	}

	return imageBuffer
}

func resizeImage(t *testing.T, imageBuffer []byte) vips.Image {
	resizedImage, err := vips.ResizeImage(imageBuffer, 500, 500)
	if err != nil {
		t.Fatal(err)
	}

	vips.SetUserComment(resizedImage, "Test")

	return resizedImage
}

func readFixture(fixtureName string, extension string) []byte {
	fixture, _ := os.ReadFile(fixturePath(fixtureName, extension))
	return fixture
}
func fixturePath(fixtureName string, extension string) string {
	return fmt.Sprintf("../../test/fixtures/vips/%s_result_%s.%s", fixtureName, runtime.GOOS, extension)
}
