package image

import (
	"io/ioutil"

	"github.com/DMarby/picsum-photos/vips"
)

// Processor is an image processor
type Processor struct {
}

// New instantiates a new image processor
func New() *Processor {
	vips.Initialize() // TODO: When should we initialize/shutdown/etc? Since that's global, but processor is instantiated...make it not instantiated?
	return &Processor{}
}

// TODO: What should we expose? Just resize, crop, etc?
func (p *Processor) LoadImage(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	image, err := vips.LoadFromBuffer(buf)
	if err != nil {
		return err
	}

	buffer, err := vips.SaveToBuffer(image)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("./fixtures/result.jpg", buffer, 0644)
	if err != nil {
		return err
	}

	return nil
}
