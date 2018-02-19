package vips

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"io/ioutil"
)

func TestImage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vips")
}

var imageBuffer []byte

var _ = BeforeSuite(func() {
	err := Initialize()
	imageBuffer, err = ioutil.ReadFile("../test/fixtures/fixture.jpg")
	Ω(err).Should(BeNil())
})

var _ = Describe("ResizeImage", func() {
	It("Loads and resizes an image", func() {
		_, err := ResizeImage(imageBuffer, 500, 500)
		Ω(err).Should(BeNil())
	})

	It("Throws an error when given an empty buffer", func() {
		var buf []byte
		_, err := ResizeImage(buf, 500, 500)
		Ω(err).Should(MatchError("empty buffer"))
	})

	It("Throws an error when given an invalid image", func() {
		_, err := ResizeImage(make([]byte, 5), 500, 500)
		Ω(err).Should(MatchError("error processing image from buffer VipsForeignLoad: buffer is not in a known format\n"))
	})
})

var _ = Describe("Grayscale", func() {
	It("Converts an image to grayscale", func() {
		image, err := ResizeImage(imageBuffer, 500, 500)
		_, err = Grayscale(image)
		Ω(err).Should(BeNil())
	})

	It("Errors on an invalid image", func() {
		_, err := Grayscale(emptyImage())
		Ω(err).Should(MatchError("error changing image colorspace vips_image_pio_input: no image data\n"))
	})
})

var _ = Describe("Blur", func() {
	It("Blurs an image", func() {
		image, err := ResizeImage(imageBuffer, 500, 500)
		_, err = Blur(image, 5)
		Ω(err).Should(BeNil())
	})

	It("Errors on an invalid image", func() {
		_, err := Blur(emptyImage(), 5)
		Ω(err).Should(MatchError("error applying blur to image vips_image_pio_input: no image data\n"))
	})
})

var _ = Describe("SaveToBuffer", func() {
	It("Saves an image to buffer", func() {
		image, err := ResizeImage(imageBuffer, 500, 500)
		_, err = SaveToBuffer(image)
		Ω(err).Should(BeNil())
	})

	It("Errors on an invalid image", func() {
		_, err := SaveToBuffer(emptyImage())
		Ω(err).Should(MatchError("error saving to buffer vips_image_pio_input: no image data\n"))
	})
})

var _ = Describe("Reproducability", func() {
	It("Produces the expected result on resize", func() {
		image, _ := ResizeImage(imageBuffer, 500, 500)
		buf, _ := SaveToBuffer(image)
		resultFixture, _ := ioutil.ReadFile("../test/fixtures/resize_result.jpg")
		Ω(buf).Should(Equal(resultFixture))
	})

	It("Produces the expected result on grayscale", func() {
		image, _ := ResizeImage(imageBuffer, 500, 500)
		image, _ = Grayscale(image)
		buf, _ := SaveToBuffer(image)
		resultFixture, _ := ioutil.ReadFile("../test/fixtures/grayscale_result.jpg")
		Ω(buf).Should(Equal(resultFixture))
	})

	It("Produces the expected result on blur", func() {
		image, _ := ResizeImage(imageBuffer, 500, 500)
		image, _ = Blur(image, 5)
		buf, _ := SaveToBuffer(image)
		resultFixture, _ := ioutil.ReadFile("../test/fixtures/blur_result.jpg")
		Ω(buf).Should(Equal(resultFixture))
	})
})

var _ = AfterSuite(func() {
	Shutdown()
})
