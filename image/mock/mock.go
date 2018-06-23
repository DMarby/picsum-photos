package mock

import (
	"fmt"

	"github.com/DMarby/picsum-photos/image"
)

// Processor implements a mock image processor
type Processor struct {
}

// ProcessImage returns an error instead of process an image
func (p *Processor) ProcessImage(task *image.Task) (processedImage []byte, err error) {
	return nil, fmt.Errorf("processing error")
}
