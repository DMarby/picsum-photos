package vips

import (
	"context"
	"fmt"
	"runtime"

	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/logger"
	"github.com/DMarby/picsum-photos/queue"
	"github.com/DMarby/picsum-photos/vips"
)

// Processor is an image processor that uses vips to process images
type Processor struct {
	queue *queue.Queue
}

// New initializes a new processor instance
func New(ctx context.Context, log *logger.Logger, cache *image.Cache) (*Processor, error) {
	err := vips.Initialize(log)
	if err != nil {
		return nil, err
	}

	workers := getWorkerCount()
	workerQueue := queue.New(ctx, workers, taskProcessor(cache))
	instance := &Processor{
		queue: workerQueue,
	}

	go workerQueue.Run()
	log.Infof("starting vips worker queue with %d workers", workers)

	return instance, err
}

func getWorkerCount() int {
	workers := runtime.NumCPU()
	return workers
}

// ProcessImage loads an image from a byte buffer, processes it, and returns a buffer containing the processed image
func (p *Processor) ProcessImage(ctx context.Context, task *image.Task) (processedImage []byte, err error) {
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

func taskProcessor(cache *image.Cache) func(data interface{}) (interface{}, error) {
	return func(data interface{}) (interface{}, error) {
		task, ok := data.(*image.Task)
		if !ok {
			return nil, fmt.Errorf("invalid data")
		}

		imageBuffer, err := cache.Get(task.ImageID)
		if err != nil {
			return nil, fmt.Errorf("error getting image from cache: %s", err)
		}

		image, err := resizeImage(imageBuffer, task.Width, task.Height)
		if err != nil {
			return nil, err
		}

		if task.ApplyBlur {
			image, err = image.blur(task.BlurAmount)
			if err != nil {
				return nil, err
			}
		}

		if task.ApplyGrayscale {
			image, err = image.grayscale()
			if err != nil {
				return nil, err
			}
		}

		buffer, err := image.saveToBuffer()
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
