package queue

import (
	"context"
	"fmt"
)

// Queue is a worker queue with a fixed amount of workers
type Queue struct {
	workers int
	queue   chan job
	handler func(interface{}) (interface{}, error)
	ctx     context.Context
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
func New(ctx context.Context, workers int, handler func(interface{}) (interface{}, error)) *Queue {
	queue := &Queue{
		workers: workers,
		queue:   make(chan job),
		handler: handler,
		ctx:     ctx,
	}

	return queue
}

// Run starts the queue and blocks until it's shut down
func (q *Queue) Run() {
	for i := 0; i < q.workers; i++ {
		go q.worker()
	}

	<-q.ctx.Done()
	close(q.queue)
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
		case <-q.ctx.Done():
			return
		}
	}
}

// Process adds a job to the queue, waits for it to process, and returns the result
func (q *Queue) Process(data interface{}) (interface{}, error) {
	if q.ctx.Err() != nil {
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
