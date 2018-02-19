package image

import "github.com/DMarby/picsum-photos/vips"

type Image struct {
	vipsImage vips.Image
}

func NewEmptyImage() *Image {
	return &Image{
		vipsImage: vips.NewEmptyImage(),
	}
}

// ResizeImage loads an image from a byte buffer, resizes it and returns an Image object for further use
func (p *Processor) ResizeImage(buffer []byte, width int, height int) (*Image, error) {
	image, err := vips.ResizeImage(buffer, width, height)

	if err != nil {
		return nil, err
	}

	return &Image{
		vipsImage: image,
	}, nil
}

// Grayscale turns an image into grayscale
func (i *Image) Grayscale() (*Image, error) {
	image, err := vips.Grayscale(i.vipsImage)
	if err != nil {
		return nil, err
	}

	return &Image{
		vipsImage: image,
	}, nil
}

// Blur applies gaussian blur to an image
func (i *Image) Blur(blur int) (*Image, error) {
	image, err := vips.Blur(i.vipsImage, blur)
	if err != nil {
		return nil, err
	}

	return &Image{
		vipsImage: image,
	}, nil
}

// SaveToBuffer returns the image as a JPEG byte buffer
func (i *Image) SaveToBuffer() ([]byte, error) {
	imageBuffer, err := vips.SaveToBuffer(i.vipsImage)

	if err != nil {
		return nil, err
	}

	return imageBuffer, nil
}
