package image

import "context"

// Processor is an image processor
type Processor interface {
	ProcessImage(ctx context.Context, task *Task) (processedImage []byte, err error)
}
