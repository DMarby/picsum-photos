package api_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/DMarby/picsum-photos/api"
	processorMock "github.com/DMarby/picsum-photos/image/mock"
	"github.com/DMarby/picsum-photos/image/vips"
	"github.com/DMarby/picsum-photos/storage"
	"github.com/DMarby/picsum-photos/storage/file"
	"github.com/DMarby/picsum-photos/storage/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "api")
}

var cancel context.CancelFunc
var router http.Handler
var mockRouter http.Handler
var mockProcessor http.Handler

var _ = BeforeSuite(func() {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	imageProcessor, _ := vips.GetInstance(ctx)
	storage, _ := file.New("../test/fixtures/file")

	router = api.New(imageProcessor, storage).Router()
	mockRouter = api.New(imageProcessor, &mock.Provider{}).Router()
	mockProcessor = api.New(&processorMock.Processor{}, storage).Router()
})

// TODO: Disable logging for requests
var _ = Describe("API", func() {
	It("/list lists images", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/list", nil)
		router.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(200))

		fixture, _ := json.Marshal([]storage.Image{
			storage.Image{
				ID:     "1",
				Author: "David Marby",
				URL:    "https://dmarby.se",
			},
		})

		Ω(w.Body.String()).Should(Equal(string(fixture) + "\n"))
	})

	DescribeTable("Errors", func(url string, code int, message string) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(code))
		Ω(w.Body.String()).Should(Equal(message + "\n"))
	},
		Entry("invalid image id", "/id/2/200", 404, "Image does not exist"),
		Entry("invalid image id", "/id/2/200/300", 404, "Image does not exist"),
		Entry("invalid size", "/id/1/0", 400, "Invalid size"),
		Entry("invalid size", "/id/1/0/0", 400, "Invalid size"),
		Entry("invalid size", "/id/1/1/9223372036854775808", 400, "Invalid size"), // Number larger then max int size to fail int parsing
		Entry("invalid size", "/id/1/9223372036854775808/1", 400, "Invalid size"), // Number larger then max int size to fail int parsing
	)

	DescribeTable("Storage Errors", func(url string, code int, message string) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		mockRouter.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(code))
		Ω(w.Body.String()).Should(Equal(message + "\n"))
	},
		Entry("List()", "/list", 500, "Something went wrong"),
		Entry("GetRandom()", "/200", 500, "Something went wrong"),
		Entry("Get()", "/id/1/100", 500, "Something went wrong"),
	)

	It("Correctly handles errors in the processor", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/id/1/100/100", nil)
		mockProcessor.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(500))
		Ω(w.Body.String()).Should(Equal("Something went wrong\n"))
	})

	DescribeTable("Images", func(url string, filePath string) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(200))
		resultFixture, _ := ioutil.ReadFile(filePath)
		Ω(w.Body.Bytes()).Should(Equal(resultFixture))
	},
		Entry("/id/:id/:size", "/id/1/200", "../test/fixtures/api/size.jpg"),
		Entry("/id/:id/:width/:height", "/id/1/200/120", "../test/fixtures/api/width_height.jpg"),
		Entry("/id/:id/:size?blur", "/id/1/200?blur", "../test/fixtures/api/blur.jpg"),
		Entry("/id/:id/:width/:height?blur", "/id/1/200/200?blur", "../test/fixtures/api/blur.jpg"),
		Entry("/id/:id/:size?grayscale", "/id/1/200?grayscale", "../test/fixtures/api/grayscale.jpg"),
		Entry("/id/:id/:width/:height?grayscale", "/id/1/200/200?grayscale", "../test/fixtures/api/grayscale.jpg"),
		Entry("/id/:id/:size?blur&grayscale", "/id/1/200?blur&grayscale", "../test/fixtures/api/all.jpg"),
		Entry("/id/:id/:width/:height?blur&grayscale", "/id/1/200/200?blur&grayscale", "../test/fixtures/api/all.jpg"),
	)

	DescribeTable("Redirects", func(url string, expectedUrl string) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(302))
		Ω(w.Header()["Location"][0]).Should(Equal(expectedUrl))
	},
		Entry("/:size", "/200", "/id/1/200/200"),
		Entry("/:width/:height", "/200/300", "/id/1/200/300"),
		Entry("/:size?grayscale", "/200?grayscale", "/id/1/200/200?grayscale"),
		Entry("/:width/:height?grayscale", "/200/300?grayscale", "/id/1/200/300?grayscale"),
		Entry("/:size?blur", "/200?blur", "/id/1/200/200?blur"),
		Entry("/:width/:height?blur", "/200/300?blur", "/id/1/200/300?blur"),
		Entry("/:size?grayscale&blur", "/200?grayscale&blur", "/id/1/200/200?grayscale&blur"),
		Entry("/:width/:height?grayscale&blur", "/200/300?grayscale&blur", "/id/1/200/300?grayscale&blur"),
		// Deprecated routes
		Entry("/g/:size", "/g/200", "/200/200?grayscale"),
		Entry("/g/:width/:height", "/g/200/300", "/200/300?grayscale"),
		Entry("/g/:size?blur", "/g/200?blur", "/200/200?grayscale&blur"),
		Entry("/g/:width/:height?blur", "/g/200/300?blur", "/200/300?grayscale&blur"),
		Entry("/g/:size?image=:id", "/g/200?image=1", "/id/1/200/200?grayscale"),
		Entry("/g/:width/:height?image=:id", "/g/200/300?image=1", "/id/1/200/300?grayscale"),
		// Deprecated query params
		Entry("/:size?image=:id", "/200?image=1", "/id/1/200/200"),
		Entry("/:width/:height?image=:id", "/200/300?image=1", "/id/1/200/300"),
		Entry("/:size?image=:id&grayscale", "/200?image=1&grayscale", "/id/1/200/200?grayscale"),
		Entry("/:width/:height?image=:id&grayscale", "/200/300?image=1&grayscale", "/id/1/200/300?grayscale"),
		Entry("/:size?image=:id&blur", "/200?image=1&blur", "/id/1/200/200?blur"),
		Entry("/:width/:height?image=:id&blur", "/200/300?image=1&blur", "/id/1/200/300?blur"),
		Entry("/:size?image=:id&grayscale&blur", "/200?image=1&grayscale&blur", "/id/1/200/200?grayscale&blur"),
		Entry("/:width/:height?image=:id&grayscale&blur", "/200/300?image=1&grayscale&blur", "/id/1/200/300?grayscale&blur"),
	)
})

var _ = AfterSuite(func() {
	cancel()
})
