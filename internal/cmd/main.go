package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Http timeouts
const (
	ReadTimeout    = 5 * time.Second
	WriteTimeout   = time.Minute
	HandlerTimeout = 45 * time.Second
)

// WaitForInterrupt waits for an interrupt
func WaitForInterrupt(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-ctx.Done():
		return errors.New("canceled")
	}
}
