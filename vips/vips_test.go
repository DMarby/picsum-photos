package vips_test

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/DMarby/picsum-photos/logger"
	"github.com/DMarby/picsum-photos/vips"
	"go.uber.org/zap"

	"testing"

	"io/ioutil"
)

func resizeImage(t *testing.T, imageBuffer []byte) vips.Image {
	resizedImage, err := vips.ResizeImage(imageBuffer, 500, 500)
	if err != nil {
		t.Fatal(err)
	}

	vips.SetUserComment(resizedImage, "Test")

	return resizedImage
}

func TestVips(t *testing.T) {
	log := logger.New(zap.FatalLevel)
	defer log.Sync()

	err := vips.Initialize(log)
	if err != nil {
		t.Fatal(err)
	}

	defer vips.Shutdown()

	imageBuffer, err := ioutil.ReadFile("../test/fixtures/fixture.jpg")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("SaveToBuffer", func(t *testing.T) {
		t.Run("saves an image to buffer", func(t *testing.T) {
			_, err := vips.SaveToBuffer(resizeImage(t, imageBuffer))
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("errors on an invalid image", func(t *testing.T) {
			_, err := vips.SaveToBuffer(vips.NewEmptyImage())
			if err == nil || err.Error() != "error saving to buffer vips_image_pio_input: no image data\n" {
				t.Error(err)
			}
		})
	})

	t.Run("ResizeImage", func(t *testing.T) {
		t.Run("loads and resizes an image", func(t *testing.T) {
			image, err := vips.ResizeImage(imageBuffer, 500, 500)
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToBuffer(image)
			resultFixture, _ := ioutil.ReadFile(fmt.Sprintf("../test/fixtures/vips/resize_result_%s.jpg", runtime.GOOS))
			if runtime.GOOS == "linux" && !reflect.DeepEqual(buf, resultFixture) {
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
		t.Run("converts an image to grayscale", func(t *testing.T) {
			image, err := vips.Grayscale(resizeImage(t, imageBuffer))
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToBuffer(image)
			resultFixture, _ := ioutil.ReadFile(fmt.Sprintf("../test/fixtures/vips/grayscale_result_%s.jpg", runtime.GOOS))
			if runtime.GOOS == "linux" && !reflect.DeepEqual(buf, resultFixture) {
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
		t.Run("blurs an image", func(t *testing.T) {
			image, err := vips.Blur(resizeImage(t, imageBuffer), 5)
			if err != nil {
				t.Error(err)
			}

			buf, _ := vips.SaveToBuffer(image)
			resultFixture, _ := ioutil.ReadFile(fmt.Sprintf("../test/fixtures/vips/blur_result_%s.jpg", runtime.GOOS))
			if runtime.GOOS == "linux" && !reflect.DeepEqual(buf, resultFixture) {
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
