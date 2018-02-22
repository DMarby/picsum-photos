package queue_test

import (
	"context"
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
var cancel context.CancelFunc

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
})

var _ = Describe("Queue", func() {
	Describe("Process", func() {
		It("Processes a task without error", func() {
			data, err := workerQueue.Process("test")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(data).Should(Equal("test"))
		})

		It("Processes a task with error", func() {
			errorQueue, errorCancel := setupQueue(func(data interface{}) (interface{}, error) {
				return nil, fmt.Errorf("custom error")
			})
			defer errorCancel()
			_, err := errorQueue.Process("test")
			Ω(err).Should(MatchError("custom error"))
		})

		It("Errors when queue is shut down", func() {
			cancel()
			_, err := workerQueue.Process("test")
			Ω(err).Should(MatchError("queue has been shutdown"))
		})
	})

})

var _ = AfterEach(func() {
	cancel()
})
