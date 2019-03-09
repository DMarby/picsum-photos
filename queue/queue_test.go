package queue_test

import (
	"context"
	"fmt"
	"testing"

	queue "github.com/DMarby/picsum-photos/queue"
)

func setupQueue(f func(data interface{}) (interface{}, error)) (*queue.Queue, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	workerQueue := queue.New(ctx, 5, f)
	go workerQueue.Run()
	return workerQueue, cancel
}

func TestProcess(t *testing.T) {
	workerQueue, cancel := setupQueue(func(data interface{}) (interface{}, error) {
		stringData, _ := data.(string)
		return stringData, nil
	})

	defer cancel()

	data, err := workerQueue.Process("test")
	if err != nil {
		t.Fatal(err)
	}

	if data != "test" {
		t.Fatal(err)
	}
}

func TestShutdown(t *testing.T) {
	workerQueue, cancel := setupQueue(func(data interface{}) (interface{}, error) {
		return "", nil
	})

	cancel()

	_, err := workerQueue.Process("test")
	if err == nil || err.Error() != "queue has been shutdown" {
		t.FailNow()
	}
}

func TestTaskWithError(t *testing.T) {
	errorQueue, cancel := setupQueue(func(data interface{}) (interface{}, error) {
		return nil, fmt.Errorf("custom error")
	})

	defer cancel()
	_, err := errorQueue.Process("test")

	if err == nil || err.Error() != "custom error" {
		t.Fatal("Invalid error")
	}
}
