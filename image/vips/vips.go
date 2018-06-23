package vips

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/queue"
	"github.com/DMarby/picsum-photos/vips"
)

// Processor is an image processor that uses vips to process images
type Processor struct {
	queue *queue.Queue
}

var instance *Processor
var once sync.Once

// GetInstance returns the processor instance, and creates it if neccesary.
func GetInstance(ctx context.Context) (*Processor, error) {
	var instance *Processor
	var err error

	once.Do(func() {
		err = vips.Initialize()
		if err != nil {
			return
		}

		workerQueue := queue.New(ctx, getWorkerCount(), instance.processImage)
		instance = &Processor{
			queue: workerQueue,
		}

		go workerQueue.Run()
	})

	return instance, err
}

func getWorkerCount() int {
	workers := runtime.NumCPU() - 1

	if workers < 1 {
		workers = 1
	}

	return workers
}

// ProcessImage loads an image from a byte buffer, processes it, and returns a buffer containing the processed image
func (p *Processor) ProcessImage(task *image.Task) (processedImage []byte, err error) {
	result, err := p.queue.Process(task)

	if err != nil {
		return nil, err
	}

	image, ok := result.([]byte)
	if !ok {
		return nil, fmt.Errorf("error getting result")
	}

	return image, nil
}

func (p *Processor) processImage(data interface{}) (interface{}, error) {
	task, ok := data.(*image.Task)
	if !ok {
		return nil, fmt.Errorf("invalid data")
	}

	image, err := p.ResizeImage(task.Buffer, task.Width, task.Height)
	if err != nil {
		return nil, err
	}

	if task.BlurAmount != 0 {
		image, err = image.Blur(task.BlurAmount)
		if err != nil {
			return nil, err
		}
	}

	if task.ApplyGrayscale {
		image, err = image.Grayscale()
		if err != nil {
			return nil, err
		}
	}

	buffer, err := image.SaveToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// Shutdown shuts down the image processor and deinitialises vips
func (p *Processor) Shutdown() {
	vips.Shutdown()
}
