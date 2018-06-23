package file_test

import (
	"io/ioutil"

	"github.com/DMarby/picsum-photos/storage"
	"github.com/DMarby/picsum-photos/storage/file"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "file")
}

var provider storage.Provider

var _ = BeforeSuite(func() {
	var err error
	provider, err = file.New("../../test/fixtures/file")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Get", func() {
	It("Gets an image by id", func() {
		buf, err := provider.Get("1")
		Ω(err).ShouldNot(HaveOccurred())
		resultFixture, _ := ioutil.ReadFile("../../test/fixtures/file/1.jpg")
		Ω(buf).Should(Equal(resultFixture))
	})

	It("Errors on a nonexistant image", func() {
		_, err := provider.Get("2")
		Ω(err).Should(MatchError(storage.ErrNotFound))
	})
})

var _ = Describe("GetRandom", func() {
	It("Returns a random image", func() {
		image, err := provider.GetRandom()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(image).Should(Equal("1"))
	})
})

var _ = Describe("List", func() {
	It("Returns a list of images", func() {
		images, err := provider.List()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(images).Should(Equal([]storage.Image{
			storage.Image{
				ID:     "1",
				Author: "David Marby",
				URL:    "https://dmarby.se",
			},
		}))
	})
})
