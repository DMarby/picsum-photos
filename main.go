package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DMarby/picsum-photos/api"
	"github.com/DMarby/picsum-photos/image/vips"
	"github.com/DMarby/picsum-photos/storage/file"
	"github.com/oklog/run"
)

func handleInterrupt(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-ctx.Done():
		return errors.New("canceled")
	}
}
func main() {
	var g run.Group

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get imageProcessor instance
	imageProcessor, err := vips.GetInstance(ctx)
	if err != nil {
		// TODO: Log, make sure this exits cleanly without canceling context or w/e
		return
	}

	// Exit if we receieve SIGINT or SIGTERM
	g.Add(func() error {
		return handleInterrupt(ctx)
	}, func(error) {
		imageProcessor.Shutdown()
	})

	// TODO: Config option? Or just always do spaces + db
	storage, err := file.New("./test/fixtures/file")
	if err != nil {
		// TODO: Log, make sure this exits everything cleanly
		cancel()
		return
	}

	// Start and listen on http
	api := api.New(imageProcessor, storage)
	server := &http.Server{
		Addr:    ":8080",
		Handler: api.Router(),
	}

	g.Add(func() error {
		return server.ListenAndServe()
	}, func(error) {
		server.Shutdown(ctx)
	})

	log.Print(g.Run().Error())
}
