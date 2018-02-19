package image_test

import (
	image "github.com/DMarby/picsum-photos/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var imageProcessor *image.Processor

func TestImage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "image")
}

var _ = BeforeSuite(func() {
	var err error
	imageProcessor, err = image.GetInstance()
	Ω(err).Should(BeNil())
})

// TODO: Clean up after each test
var _ = Describe("Image", func() {
	It("Should do something", func() {
		err := imageProcessor.LoadImage("../test/fixtures/fixture.jpg")
		Ω(err).Should(BeNil())
	})

	Measure("Should do something perf", func(b Benchmarker) {
		b.Time("runtime", func() {
			err := imageProcessor.LoadImage("../test/fixtures/fixture.jpg")
			Ω(err).Should(BeNil())
		})
	}, 10)
})

var _ = AfterSuite(func() {
	imageProcessor.Shutdown()
})
