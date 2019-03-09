package vips_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"

	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/image/vips"
	"github.com/DMarby/picsum-photos/logger"
	"go.uber.org/zap"

	"testing"
)

func setup() (context.CancelFunc, *vips.Processor, []byte, error) {
	log := logger.New(zap.ErrorLevel)
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())

	processor, err := vips.GetInstance(ctx, log)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	buf, err := ioutil.ReadFile("../../test/fixtures/fixture.jpg")
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	return cancel, processor, buf, nil
}

func fullTest(processor *vips.Processor, buf []byte) []byte {
	task := image.NewTask(buf, 500, 500).Grayscale().Blur(5)
	imageBuffer, _ := processor.ProcessImage(task)
	return imageBuffer
}

func TestVips(t *testing.T) {
	cancel, processor, buf, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	defer processor.Shutdown()

	t.Run("Processor", func(t *testing.T) {
		t.Run("resize image", func(t *testing.T) {
			_, err := processor.ResizeImage(buf, 500, 500)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("resize image handles errors", func(t *testing.T) {
			_, err := processor.ResizeImage(make([]byte, 0), 500, 500)
			if err == nil || err.Error() != "empty buffer" {
				t.Error()
			}
		})

		t.Run("process image", func(t *testing.T) {
			_, err := processor.ProcessImage(image.NewTask(buf, 500, 500))
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("process image handles errors", func(t *testing.T) {
			_, err := processor.ProcessImage(image.NewTask(make([]byte, 0), 500, 500))
			if err == nil || err.Error() != "empty buffer" {
				t.Error()
			}
		})

		t.Run("full test", func(t *testing.T) {
			resultFixture, _ := ioutil.ReadFile(fmt.Sprintf("../../test/fixtures/image/complete_result_%s.jpg", runtime.GOOS))
			testResult := fullTest(processor, buf)
			if !reflect.DeepEqual(testResult, resultFixture) {
				t.Error("image data doesn't match")
			}
		})
	})

	t.Run("Image", func(t *testing.T) {
		resizedImage, err := processor.ResizeImage(buf, 500, 500)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("converts an image to grayscale", func(t *testing.T) {
			_, err := resizedImage.Grayscale()
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("grayscale handles errors", func(t *testing.T) {
			testImage := vips.NewEmptyImage()
			_, err := testImage.Grayscale()
			if err == nil || err.Error() != "error changing image colorspace vips_image_pio_input: no image data\n" {
				t.Error("wrong error")
			}
		})

		t.Run("blurs an image", func(t *testing.T) {
			_, err := resizedImage.Blur(5)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("blur handles errors", func(t *testing.T) {
			testImage := vips.NewEmptyImage()
			_, err := testImage.Blur(5)
			if err == nil || err.Error() != "error applying blur to image vips_image_pio_input: no image data\n" {
				t.Error("wrong error")
			}
		})

		t.Run("save to buffer", func(t *testing.T) {
			_, err := resizedImage.SaveToBuffer()
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("save to buffer handles errors", func(t *testing.T) {
			testImage := vips.NewEmptyImage()
			_, err := testImage.SaveToBuffer()
			if err == nil || err.Error() != "error saving to buffer vips_image_pio_input: no image data\n" {
				t.Error("wrong error")
			}
		})
	})
}

func BenchmarkVips(b *testing.B) {
	cancel, processor, buf, err := setup()
	if err != nil {
		b.Fatal(err)
	}
	defer cancel()
	defer processor.Shutdown()

	resizedImage, err := processor.ResizeImage(buf, 500, 500)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("full test", func(b *testing.B) {
		fullTest(processor, buf)
	})

	b.Run("resizeImage", func(b *testing.B) {
		processor.ResizeImage(buf, 500, 500)
	})

	b.Run("grayscale", func(b *testing.B) {
		resizedImage.Grayscale()
	})

	b.Run("blur", func(b *testing.B) {
		resizedImage.Blur(5)
	})

	b.Run("saveToBuffer", func(b *testing.B) {
		resizedImage.SaveToBuffer()
	})
}
