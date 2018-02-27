package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/DMarby/picsum-photos/api"
	"github.com/DMarby/picsum-photos/queue"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestImage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "api")
}

var workerQueue *queue.Queue
var cancel context.CancelFunc
var router http.Handler

func setupQueue(f func(data interface{}) (interface{}, error)) (*queue.Queue, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	workerQueue := queue.New(ctx, 5, f)
	go workerQueue.Run()
	return workerQueue, cancel
}

var _ = BeforeEach(func() {
	workerQueue, cancel = setupQueue(func(data interface{}) (interface{}, error) {
		stringData, _ := data.(string)
		return stringData, nil
	})

	router = api.New(workerQueue).Router()
})

// TODO: Disable logging for requests
var _ = Describe("API", func() {
	It("/list lists images", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/list", nil)
		router.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(200))
		Ω(w.Body.String()).Should(Equal("[]\n")) // TODO: Why is a newline added?
	})

	It("/:size returns square images of the correct size", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/5", nil)
		router.ServeHTTP(w, req)
		Ω(w.Code).Should(Equal(200))
		Ω(w.Body.String()).Should(Equal("5"))
	})
})

var _ = AfterEach(func() {
	cancel()
})
