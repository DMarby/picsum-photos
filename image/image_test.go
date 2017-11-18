package image_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestImage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "image")
}

var _ = Describe("Image", func() {
	It("Should do something", func() {
		Ω(true).Should(Equal(true))
	})
})
