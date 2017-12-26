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

// TODO: Benchmarks
var _ = Describe("Image", func() {
	It("Should do something", func() {
		Ω(true).Should(Equal(true))
		imageProcessor := image.New()
		err := imageProcessor.LoadImage("./fixtures/fixture.jpg")
		Ω(err).Should(BeNil())
	})
})
