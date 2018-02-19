package image

import (
	"sync"

	"github.com/DMarby/picsum-photos/vips"
)

// Processor is an image processor
type Processor struct {
}

var instance *Processor
var once sync.Once

// GetInstance returns the processor instance, and creates it if neccesary.
func GetInstance() (*Processor, error) {
	var err error

	once.Do(func() {
		err = vips.Initialize()
		instance = &Processor{}
	})

	return instance, err
}

// Shutdown shuts down the image processor and deinitialises vips
func (p *Processor) Shutdown() {
	vips.Shutdown()
}
