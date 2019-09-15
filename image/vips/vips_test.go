package vips_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"

	"github.com/DMarby/picsum-photos/cache/memory"
	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/image/vips"
	"github.com/DMarby/picsum-photos/logger"
	"github.com/DMarby/picsum-photos/storage/file"
	"go.uber.org/zap"

	"testing"
)

func setup() (context.CancelFunc, *vips.Processor, []byte, error) {
	log := logger.New(zap.ErrorLevel)
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	storage, err := file.New("../../test/fixtures/file")
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	cache := image.NewCache(memory.New(), storage)

	processor, err := vips.New(ctx, log, cache)
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
	task := image.NewTask("1", 500, 500, "testing").Grayscale().Blur(5)
	imageBuffer, _ := processor.ProcessImage(context.Background(), task)
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
		t.Run("process image", func(t *testing.T) {
			_, err := processor.ProcessImage(context.Background(), image.NewTask("1", 500, 500, "testing"))
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("process image handles errors", func(t *testing.T) {
			_, err := processor.ProcessImage(context.Background(), image.NewTask("foo", 500, 500, "testing"))
			if err == nil || err.Error() != "error getting image from cache: open ../../test/fixtures/file/foo.jpg: no such file or directory" {
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
}

func BenchmarkVips(b *testing.B) {
	cancel, processor, buf, err := setup()
	if err != nil {
		b.Fatal(err)
	}
	defer cancel()
	defer processor.Shutdown()

	b.Run("full test", func(b *testing.B) {
		fullTest(processor, buf)
	})
}
