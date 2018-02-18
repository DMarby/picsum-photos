package image_test

import (
	image "github.com/DMarby/picsum-photos/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestImage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "image")
}

// TODO: Just re-use one imageProcessor
// TODO: Clean up after each test
var _ = Describe("Image", func() {
	It("Should do something", func() {
		imageProcessor := image.New()
		err := imageProcessor.LoadImage("./fixtures/fixture.jpg")
		Ω(err).Should(BeNil())
	})

	Measure("Should do something perf", func(b Benchmarker) {
		imageProcessor := image.New()

		runtime := b.Time("runtime", func() {
			err := imageProcessor.LoadImage("./fixtures/fixture.jpg")
			Ω(err).Should(BeNil())
		})

		Ω(runtime.Seconds()).Should(BeNumerically("<", 0.2), "ProcessImage() shouldn't take too long.")
	}, 10)
})
