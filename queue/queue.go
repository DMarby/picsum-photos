package queue

import (
	"fmt"
	"sync"
)

// Queue is a worker queue with a fixed amount of workers
type Queue struct {
	queue   chan job
	handler func(interface{}) (interface{}, error)
	once    sync.Once
}

type job struct {
	data   interface{}
	result chan jobResult
}

type jobResult struct {
	result interface{}
	err    error
}

// New creates a new Queue with the specified amount of workers
func New(workers int, handler func(interface{}) (interface{}, error)) *Queue {
	queue := &Queue{
		queue:   make(chan job),
		handler: handler,
	}

	for i := 0; i < workers; i++ {
		go queue.worker()
	}

	return queue
}

func (q *Queue) worker() {
	for {
		select {
		case job, open := <-q.queue:
			if !open {
				return
			}

			result, err := q.handler(job.data)
			job.result <- jobResult{
				result: result,
				err:    err,
			}
		}
	}
}

// Process adds a job to the queue, waits for it to process, and returns the result
func (q *Queue) Process(data interface{}) (interface{}, error) {
	if q.queue == nil {
		return nil, fmt.Errorf("queue has been shutdown")
	}

	resultChan := make(chan jobResult)

	q.queue <- job{
		data:   data,
		result: resultChan,
	}

	result := <-resultChan
	close(resultChan)

	if result.err != nil {
		return nil, result.err
	}

	return result.result, nil
}

// Shutdown shuts down the queue after all currently running tasks are finished
func (q *Queue) Shutdown() {
	q.once.Do(func() {
		close(q.queue)
		q.queue = nil
	})
}
