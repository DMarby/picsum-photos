package image

import (
	"io/ioutil"

	"github.com/DMarby/picsum-photos/vips"
)

// Processor is an image processor
type Processor struct {
}

type Image struct {
	data []byte
}

// New instantiates a new image processor and initializes vips
func New() *Processor {
	vips.Initialize()
	return &Processor{}
}

// Shutdown shuts down the image processor and deinitialises vips
func (p *Processor) Shutdown() {
	vips.Shutdown()
}

// TODO: What should we expose? Just resize, crop, etc?
func (p *Processor) LoadImage(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	image, err := vips.ProcessImage(buf)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("./fixtures/result.jpg", image, 0644)
	if err != nil {
		return err
	}

	return nil
}
