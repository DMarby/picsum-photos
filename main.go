package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DMarby/picsum-photos/api"
	fileDatabase "github.com/DMarby/picsum-photos/database/file"
	"github.com/DMarby/picsum-photos/health"
	vipsProcessor "github.com/DMarby/picsum-photos/image/vips"
	"github.com/DMarby/picsum-photos/logger"
	fileStorage "github.com/DMarby/picsum-photos/storage/file"
	"github.com/jamiealquiza/envy"
	"go.uber.org/zap"
)

func waitForInterrupt(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-ctx.Done():
		return errors.New("canceled")
	}
}

const readTimeout = 5 * time.Second
const writeTimeout = 30 * time.Second
const maxImageSize = 5000 // The max allowed image width/height to be requested

func main() {
	// Set up commandline flags
	listen := flag.String("listen", ":8080", "listen address")
	rootURL := flag.String("root_url", "https://picsum.photos", "root url")
	loglevel := zap.LevelFlag("log_level", zap.InfoLevel, "log level (default \"info\") (debug, info, warn, error, dpanic, panic, fatal)")

	// Parse environment variables
	envy.Parse("PICSUM")

	// Parse commandline flags
	flag.Parse()

	// Initialize the logger
	log := logger.New(*loglevel)
	defer log.Sync()

	// Set up context for shutting down
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	// Get imageProcessor instance
	imageProcessorCtx, imageProcessorCancel := context.WithCancel(context.Background())
	defer imageProcessorCancel()

	imageProcessor, err := vipsProcessor.GetInstance(imageProcessorCtx, log)
	if err != nil {
		log.Fatalf("error initializing image processor %s", err.Error())
	}

	// Initialize the storage
	storage, err := fileStorage.New("./test/fixtures/file")
	if err != nil {
		log.Fatalf("error initializing storage %s", err.Error())
		return
	}

	// Initialize the database
	database, err := fileDatabase.New("./test/fixtures/file/metadata.json")
	if err != nil {
		log.Fatalf("error initializing database %s", err.Error())
		return
	}

	// Initialize and start the health checker
	checkerCtx, checkerCancel := context.WithCancel(context.Background())
	defer checkerCancel()

	checker := health.New(checkerCtx, imageProcessor, storage, database)
	go checker.Run()

	// Start and listen on http
	api := &api.API{
		ImageProcessor: imageProcessor,
		Storage:        storage,
		Database:       database,
		HealthChecker:  checker,
		Log:            log,
		MaxImageSize:   maxImageSize,
		RootURL:        *rootURL,
	}
	server := &http.Server{
		Addr:         *listen,
		Handler:      api.Router(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Infof("shutting down the http server: %s", err)
			shutdown()
		}
	}()

	log.Infof("http server listening on %s", *listen)

	// Wait for shutdown or error
	err = waitForInterrupt(shutdownCtx)
	log.Infof("shutting down: %s", err)

	// Shut down http server
	serverCtx, serverCancel := context.WithTimeout(context.Background(), writeTimeout)
	defer serverCancel()
	if err := server.Shutdown(serverCtx); err != nil {
		log.Warnf("error shutting down: %s", err)
	}
}
