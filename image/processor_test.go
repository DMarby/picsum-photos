package image_test

import (
	"io/ioutil"

	image "github.com/DMarby/picsum-photos/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestImage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "image")
}

var buf []byte
var processor *image.Processor

var _ = BeforeSuite(func() {
	var err error
	processor, err = image.GetInstance()
	Ω(err).Should(BeNil())
	buf, err = ioutil.ReadFile("../test/fixtures/fixture.jpg")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Processor", func() {
	Describe("ResizeImage", func() {
		It("Resizes an image", func() {
			_, err := processor.ResizeImage(buf, 500, 500)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Handles errors correctly", func() {
			_, err := processor.ResizeImage(make([]byte, 0), 500, 500)
			Ω(err).Should(MatchError("empty buffer"))
		})

		Measure("Performs", func(b Benchmarker) {
			b.Time("runtime", func() {
				_, err := processor.ResizeImage(buf, 500, 500)
				Ω(err).ShouldNot(HaveOccurred())
			})
		}, 10)
	})
})

var _ = Describe("Image", func() {
	var resizedImage *image.Image

	var _ = BeforeEach(func() {
		var err error
		resizedImage, err = processor.ResizeImage(buf, 500, 500)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Describe("Grayscale", func() {
		It("Converts the image to grayscale", func() {
			_, err := resizedImage.Grayscale()
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Handles errors correctly", func() {
			testImage := image.NewEmptyImage()
			_, err := testImage.Grayscale()
			Ω(err).Should(MatchError("error changing image colorspace vips_image_pio_input: no image data\n"))
		})

		Measure("Performs", func(b Benchmarker) {
			b.Time("runtime", func() {
				_, err := resizedImage.Grayscale()
				Ω(err).ShouldNot(HaveOccurred())
			})
		}, 10)
	})

	Describe("Blur", func() {
		It("Applies gaussian blur to the image", func() {
			_, err := resizedImage.Blur(5)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Handles errors correctly", func() {
			testImage := image.NewEmptyImage()
			_, err := testImage.Blur(5)
			Ω(err).Should(MatchError("error applying blur to image vips_image_pio_input: no image data\n"))
		})

		Measure("Performs", func(b Benchmarker) {
			b.Time("runtime", func() {
				_, err := resizedImage.Blur(5)
				Ω(err).ShouldNot(HaveOccurred())
			})
		}, 10)
	})

	Describe("SaveToBuffer", func() {
		It("Saves the image to a buffer", func() {
			_, err := resizedImage.SaveToBuffer()
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Handles errors correctly", func() {
			testImage := image.NewEmptyImage()
			_, err := testImage.SaveToBuffer()
			Ω(err).Should(MatchError("error saving to buffer vips_image_pio_input: no image data\n"))
		})

		Measure("Performs", func(b Benchmarker) {
			b.Time("runtime", func() {
				_, err := resizedImage.SaveToBuffer()
				Ω(err).ShouldNot(HaveOccurred())
			})
		}, 10)
	})
})

func fullTest() {
	image, _ := processor.ResizeImage(buf, 500, 500)
	image, _ = image.Grayscale()
	image, _ = image.Blur(5)
	imageBuffer, _ := image.SaveToBuffer()
	resultFixture, _ := ioutil.ReadFile("../test/fixtures/complete_result.jpg")
	Ω(imageBuffer).Should(Equal(resultFixture))
}

var _ = Describe("Full test", func() {
	It("Produces the expected result", func() {
		fullTest()
	})

	Measure("Performs", func(b Benchmarker) {
		b.Time("runtime", func() {
			fullTest()
		})
	}, 10)
})

var _ = AfterSuite(func() {
	processor.Shutdown()
})
