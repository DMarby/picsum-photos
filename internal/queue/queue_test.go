package queue_test

import (
	"context"
	"fmt"
	"testing"

	queue "github.com/DMarby/picsum-photos/internal/queue"
)

func setupQueue(f func(ctx context.Context, data interface{}) (interface{}, error)) (*queue.Queue, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	workerQueue := queue.New(ctx, 5, f)
	go workerQueue.Run()
	return workerQueue, cancel
}

func TestProcess(t *testing.T) {
	workerQueue, cancel := setupQueue(func(ctx context.Context, data interface{}) (interface{}, error) {
		stringData, _ := data.(string)
		return stringData, nil
	})

	defer cancel()

	data, err := workerQueue.Process(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	if data != "test" {
		t.Fatal(err)
	}
}

func TestShutdown(t *testing.T) {
	workerQueue, cancel := setupQueue(func(ctx context.Context, data interface{}) (interface{}, error) {
		return "", nil
	})

	cancel()

	_, err := workerQueue.Process(context.Background(), "test")
	if err == nil || err.Error() != "queue has been shutdown" {
		t.FailNow()
	}
}

func TestTaskWithError(t *testing.T) {
	errorQueue, cancel := setupQueue(func(ctx context.Context, data interface{}) (interface{}, error) {
		return nil, fmt.Errorf("custom error")
	})

	defer cancel()
	_, err := errorQueue.Process(context.Background(), "test")

	if err == nil || err.Error() != "custom error" {
		t.Fatal("Invalid error")
	}
}

func TestTaskWithCancelledContext(t *testing.T) {
	errorQueue, cancel := setupQueue(func(ctx context.Context, data interface{}) (interface{}, error) {
		return nil, fmt.Errorf("custom error")
	})

	defer cancel()

	ctx, ctxCancel := context.WithCancel(context.Background())
	ctxCancel()

	_, err := errorQueue.Process(ctx, "test")

	if err == nil || err.Error() != "context canceled" {
		t.Fatal("Invalid error")
	}
}
