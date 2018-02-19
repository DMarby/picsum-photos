package vips_test

import (
	"github.com/DMarby/picsum-photos/vips"
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
	err := vips.Initialize()
	Ω(err).ShouldNot(HaveOccurred())
	imageBuffer, err = ioutil.ReadFile("../test/fixtures/fixture.jpg")
	Ω(err).ShouldNot(HaveOccurred())
})

var resizedImage vips.Image
var emptyImage vips.Image

var _ = BeforeEach(func() {
	var err error
	resizedImage, err = vips.ResizeImage(imageBuffer, 500, 500)
	Ω(err).ShouldNot(HaveOccurred())
	emptyImage = vips.NewEmptyImage()
})

var _ = Describe("SaveToBuffer", func() {
	It("Saves an image to buffer", func() {
		_, err := vips.SaveToBuffer(resizedImage)
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("Errors on an invalid image", func() {
		_, err := vips.SaveToBuffer(emptyImage)
		Ω(err).Should(MatchError("error saving to buffer vips_image_pio_input: no image data\n"))
	})
})

var _ = Describe("ResizeImage", func() {
	It("Loads and resizes an image", func() {
		image, err := vips.ResizeImage(imageBuffer, 500, 500)
		Ω(err).ShouldNot(HaveOccurred())
		buf, _ := vips.SaveToBuffer(image)
		resultFixture, _ := ioutil.ReadFile("../test/fixtures/resize_result.jpg")
		Ω(buf).Should(Equal(resultFixture))
	})

	It("Throws an error when given an empty buffer", func() {
		var buf []byte
		_, err := vips.ResizeImage(buf, 500, 500)
		Ω(err).Should(MatchError("empty buffer"))
	})

	It("Throws an error when given an invalid image", func() {
		_, err := vips.ResizeImage(make([]byte, 5), 500, 500)
		Ω(err).Should(MatchError("error processing image from buffer VipsForeignLoad: buffer is not in a known format\n"))
	})
})

var _ = Describe("Grayscale", func() {
	It("Converts an image to grayscale", func() {
		image, err := vips.Grayscale(resizedImage)
		Ω(err).ShouldNot(HaveOccurred())
		buf, _ := vips.SaveToBuffer(image)
		resultFixture, _ := ioutil.ReadFile("../test/fixtures/grayscale_result.jpg")
		Ω(buf).Should(Equal(resultFixture))
	})

	It("Errors on an invalid image", func() {
		_, err := vips.Grayscale(emptyImage)
		Ω(err).Should(MatchError("error changing image colorspace vips_image_pio_input: no image data\n"))
	})
})

var _ = Describe("Blur", func() {
	It("Blurs an image", func() {
		image, err := vips.Blur(resizedImage, 5)
		Ω(err).ShouldNot(HaveOccurred())
		buf, _ := vips.SaveToBuffer(image)
		resultFixture, _ := ioutil.ReadFile("../test/fixtures/blur_result.jpg")
		Ω(buf).Should(Equal(resultFixture))
	})

	It("Errors on an invalid image", func() {
		_, err := vips.Blur(emptyImage, 5)
		Ω(err).Should(MatchError("error applying blur to image vips_image_pio_input: no image data\n"))
	})
})

var _ = Describe("PrintDebugInfo", func() {
	It("Prints debug info", func() {
		vips.PrintDebugInfo()
	})
})

var _ = AfterSuite(func() {
	vips.Shutdown()
})
