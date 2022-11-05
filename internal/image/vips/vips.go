package vips

import (
	"context"
	"expvar"
	"fmt"
	"math"

	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/queue"
	"github.com/DMarby/picsum-photos/internal/vips"
)

// Processor is an image processor that uses vips to process images
type Processor struct {
	queue *queue.Queue
}

var (
	queueSize       = expvar.NewInt("gauge_image_processor_queue_size")
	processedImages = expvar.NewMap("counter_labelmap_dimensions_image_processor_processed_images")
)

// New initializes a new processor instance
func New(ctx context.Context, log *logger.Logger, workers int, cache *image.Cache) (*Processor, error) {
	err := vips.Initialize(log)
	if err != nil {
		return nil, err
	}

	workerQueue := queue.New(ctx, workers, taskProcessor(cache))
	instance := &Processor{
		queue: workerQueue,
	}

	go workerQueue.Run()
	log.Infof("starting vips worker queue with %d workers", workers)

	return instance, err
}

// ProcessImage loads an image from a byte buffer, processes it, and returns a buffer containing the processed image
func (p *Processor) ProcessImage(ctx context.Context, task *image.Task) (processedImage []byte, err error) {
	queueSize.Add(1)
	defer queueSize.Add(-1)

	defer processedImages.Add(fmt.Sprintf("%0.f", math.Max(math.Round(float64(task.Width)/500)*500, math.Round(float64(task.Height)/500)*500)), 1)

	result, err := p.queue.Process(ctx, task)
	if err != nil {
		return nil, err
	}

	image, ok := result.([]byte)
	if !ok {
		return nil, fmt.Errorf("error getting result")
	}

	return image, nil
}

func taskProcessor(cache *image.Cache) func(ctx context.Context, data interface{}) (interface{}, error) {
	return func(ctx context.Context, data interface{}) (interface{}, error) {
		task, ok := data.(*image.Task)
		if !ok {
			return nil, fmt.Errorf("invalid data")
		}

		// Use a pre-processed source image closer to the desired size then the original
		imageKey := task.ImageID
		width := math.Ceil(float64(task.Width)/500) * 500
		height := math.Ceil(float64(task.Height)/500) * 500
		size := math.Max(width, height)
		if size <= 4500 { // Files larger then 4500 doesn't have a suffix
			imageKey = fmt.Sprintf("%s_%0.f", task.ImageID, size)
		}

		imageBuffer, err := cache.Get(ctx, imageKey)
		if err != nil {
			return nil, fmt.Errorf("error getting image from cache: %s", err)
		}

		processedImage, err := resizeImage(imageBuffer, task.Width, task.Height)
		if err != nil {
			return nil, err
		}

		if task.ApplyBlur {
			processedImage, err = processedImage.blur(task.BlurAmount)
			if err != nil {
				return nil, err
			}
		}

		if task.ApplyGrayscale {
			processedImage, err = processedImage.grayscale()
			if err != nil {
				return nil, err
			}
		}

		processedImage.setUserComment(task.UserComment)

		var buffer []byte
		switch task.OutputFormat {
		case image.JPEG:
			buffer, err = processedImage.saveToJpegBuffer()
		case image.WebP:
			buffer, err = processedImage.saveToWebPBuffer()
		}

		if err != nil {
			return nil, err
		}

		return buffer, nil
	}
}

// Shutdown shuts down the image processor and deinitialises vips
func (p *Processor) Shutdown() {
	vips.Shutdown()
}
