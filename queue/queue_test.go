package queue_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	queue "github.com/DMarby/picsum-photos/queue"

	"testing"
)

func TestImage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "queue")
}

var workerQueue *queue.Queue

var _ = BeforeEach(func() {
	workerQueue = queue.New(5, func(data interface{}) (interface{}, error) {
		stringData := data.(string)
		return stringData, nil
	})
})

var _ = Describe("Queue", func() {
	Describe("Process", func() {
		It("Processes a task without error", func() {
			data, err := workerQueue.Process("test")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(data).Should(Equal("test"))
		})

		It("Processes a task with error", func() {
			errorQueue := queue.New(5, func(data interface{}) (interface{}, error) {
				return nil, fmt.Errorf("custom error")
			})
			_, err := errorQueue.Process("test")
			Ω(err).Should(MatchError("custom error"))
			errorQueue.Shutdown()
		})

		It("Errors when queue is shut down", func() {
			workerQueue.Shutdown()
			_, err := workerQueue.Process("test")
			Ω(err).Should(MatchError("queue has been shutdown"))
		})
	})

	Describe("Shutdown", func() {
		It("Can run twice", func() {
			workerQueue.Shutdown()
			workerQueue.Shutdown()
		})
	})
})

var _ = AfterEach(func() {
	workerQueue.Shutdown()
})
