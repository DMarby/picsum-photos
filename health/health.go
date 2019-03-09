package health

import (
	"context"
	"sync"
	"time"

	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/storage"
)

const checkInterval = 30 * time.Second
const checkTimeout = 20 * time.Second

// Checker is a periodic health checker
type Checker struct {
	ctx            context.Context
	status         Status
	imageProcessor image.Processor
	storage        storage.Provider
	database       database.Provider
	mutex          sync.RWMutex
}

// Status contains the healtcheck status
type Status struct {
	Healthy   bool   `json:"healthy"`
	Database  string `json:"database"`
	Processor string `json:"processor"`
	Storage   string `json:"storage"`
}

// New creates and returns a new health checker
func New(ctx context.Context, imageProcessor image.Processor, storage storage.Provider, database database.Provider) *Checker {
	return &Checker{
		ctx:            ctx,
		imageProcessor: imageProcessor,
		storage:        storage,
		database:       database,
	}
}

// Run starts the health checker
func (c *Checker) Run() {
	ticker := time.NewTicker(checkInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.runCheck()
			case <-c.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	c.runCheck()
}

// Status returns the status of the health checks
func (c *Checker) Status() Status {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.status
}

func (c *Checker) runCheck() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	channel := make(chan Status, 1)
	go func() {
		c.check(channel)
	}()

	select {
	case <-ctx.Done():
		c.status = Status{
			Healthy:   false,
			Database:  "unknown",
			Processor: "unknown",
			Storage:   "unknown",
		}
	case status := <-channel:
		c.status = status
	}
}

func (c *Checker) check(channel chan Status) {
	defer close(channel)

	status := Status{
		Healthy:   true,
		Database:  "unknown",
		Storage:   "unknown",
		Processor: "unknown",
	}

	id, err := c.database.GetRandom()
	if err != nil {
		status.Healthy = false
		status.Database = "unhealthy"
		channel <- status
		return
	}
	status.Database = "healthy"

	buf, err := c.storage.Get(id)
	if err != nil {
		status.Healthy = false
		status.Storage = "unhealthy"
		channel <- status
		return
	}
	status.Storage = "healthy"

	task := image.NewTask(buf, 1, 1)
	_, err = c.imageProcessor.ProcessImage(task)
	if err != nil {
		status.Healthy = false
		status.Processor = "unhealthy"
	} else {
		status.Processor = "healthy"
	}

	channel <- status
}
