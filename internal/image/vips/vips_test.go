package vips_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"

	"github.com/DMarby/picsum-photos/internal/cache/memory"
	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/image/vips"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/storage/file"
	"go.uber.org/zap"

	"testing"
)

var (
	jpegFixture = fmt.Sprintf("../../../test/fixtures/image/complete_result_%s.jpg", runtime.GOOS)
	webpFixture = fmt.Sprintf("../../../test/fixtures/image/complete_result_%s.webp", runtime.GOOS)
)

func TestVips(t *testing.T) {
	cancel, processor, buf, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	defer processor.Shutdown()

	t.Run("Processor", func(t *testing.T) {
		t.Run("process image", func(t *testing.T) {
			_, err := processor.ProcessImage(context.Background(), image.NewTask("1", 500, 500, "testing", image.JPEG))
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("process image handles errors", func(t *testing.T) {
			_, err := processor.ProcessImage(context.Background(), image.NewTask("foo", 500, 500, "testing", image.JPEG))
			if err == nil || err.Error() != "error getting image from cache: Image does not exist" {
				t.Error()
			}
		})

		t.Run("full test jpeg", func(t *testing.T) {
			resultFixture, _ := os.ReadFile(jpegFixture)
			testResult := fullTest(processor, buf, image.JPEG)
			if !reflect.DeepEqual(testResult, resultFixture) {
				t.Error("image data doesn't match")
			}
		})

		t.Run("full test webp", func(t *testing.T) {
			resultFixture, _ := os.ReadFile(webpFixture)
			testResult := fullTest(processor, buf, image.WebP)
			if !reflect.DeepEqual(testResult, resultFixture) {
				t.Error("image data doesn't match")
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

	b.Run("full test jpeg", func(b *testing.B) {
		fullTest(processor, buf, image.JPEG)
	})

	b.Run("full test webp", func(b *testing.B) {
		fullTest(processor, buf, image.WebP)
	})
}

// Utility function for regenerating the fixtures
func TestFixtures(t *testing.T) {
	if os.Getenv("GENERATE_FIXTURES") != "1" {
		t.SkipNow()
	}

	cancel, processor, buf, err := setup()
	if err != nil {
		t.Fatal(err)
	}

	defer cancel()
	defer processor.Shutdown()

	jpeg := fullTest(processor, buf, image.JPEG)
	os.WriteFile(jpegFixture, jpeg, 0644)

	webp := fullTest(processor, buf, image.WebP)
	os.WriteFile(webpFixture, webp, 0644)
}

func setup() (context.CancelFunc, *vips.Processor, []byte, error) {
	log := logger.New(zap.ErrorLevel)
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	storage, err := file.New("../../../test/fixtures/file")
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	cache := image.NewCache(memory.New(), storage)

	processor, err := vips.New(ctx, log, 3, cache)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	buf, err := os.ReadFile("../../../test/fixtures/fixture.jpg")
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	return cancel, processor, buf, nil
}

func fullTest(processor *vips.Processor, buf []byte, format image.OutputFormat) []byte {
	task := image.NewTask("1", 500, 500, "testing", format).Grayscale().Blur(5)
	imageBuffer, _ := processor.ProcessImage(context.Background(), task)
	return imageBuffer
}
