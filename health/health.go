package health

import (
	"context"
	"sync"
	"time"

	"github.com/DMarby/picsum-photos/cache"
	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/logger"
	"github.com/DMarby/picsum-photos/storage"
)

const checkInterval = 30 * time.Second
const checkTimeout = 20 * time.Second

// Checker is a periodic health checker
type Checker struct {
	Ctx      context.Context
	Storage  storage.Provider
	Database database.Provider
	Cache    cache.Provider
	status   Status
	mutex    sync.RWMutex
	Log      *logger.Logger
}

// Status contains the healtcheck status
type Status struct {
	Healthy  bool   `json:"healthy"`
	Cache    string `json:"cache"`
	Database string `json:"database"`
	Storage  string `json:"storage"`
}

// Run starts the health checker
func (c *Checker) Run() {
	ticker := time.NewTicker(checkInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.runCheck()
			case <-c.Ctx.Done():
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
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	channel := make(chan Status, 1)
	go func() {
		c.check(channel)
	}()

	select {
	case <-ctx.Done():
		c.mutex.Lock()
		c.status = Status{
			Healthy:  false,
			Cache:    "unknown",
			Database: "unknown",
			Storage:  "unknown",
		}
		c.mutex.Unlock()
		c.Log.Errorw("healthcheck timed out")
	case status := <-channel:
		c.mutex.Lock()
		c.status = status
		c.mutex.Unlock()
		if !status.Healthy {
			c.Log.Errorw("healthcheck error",
				"status", status,
			)
		}
	}
}

func (c *Checker) check(channel chan Status) {
	defer close(channel)

	status := Status{
		Healthy:  true,
		Cache:    "unknown",
		Database: "unknown",
		Storage:  "unknown",
	}

	if _, err := c.Cache.Get("healthcheck"); err != cache.ErrNotFound {
		status.Healthy = false
		status.Cache = "unhealthy"
	} else {
		status.Cache = "healthy"
	}

	id, err := c.Database.GetRandom()
	if err != nil {
		status.Healthy = false
		status.Database = "unhealthy"
		channel <- status
		return
	}
	status.Database = "healthy"

	_, err = c.Storage.Get(context.Background(), id)
	if err != nil {
		status.Healthy = false
		status.Storage = "unhealthy"
		channel <- status
		return
	}
	status.Storage = "healthy"

	channel <- status
}
