package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DMarby/picsum-photos/internal/api"
	"github.com/DMarby/picsum-photos/internal/cmd"
	"github.com/DMarby/picsum-photos/internal/hmac"
	"github.com/DMarby/picsum-photos/internal/metrics"
	"github.com/DMarby/picsum-photos/internal/tracing"

	fileDatabase "github.com/DMarby/picsum-photos/internal/database/file"
	"github.com/DMarby/picsum-photos/internal/health"
	"github.com/DMarby/picsum-photos/internal/logger"

	"github.com/jamiealquiza/envy"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

const (
	staticPath = "./dist" // Path where the static files are located
)

// Comandline flags
var (
	// Global
	listen          = flag.String("listen", ":8080", "listen address")
	metricsListen   = flag.String("metrics-listen", ":8082", "metrics listen address")
	rootURL         = flag.String("root-url", "https://picsum.photos", "root url")
	imageServiceURL = flag.String("image-service-url", "https://i.picsum.photos", "image service url")
	loglevel        = zap.LevelFlag("log-level", zap.InfoLevel, "log level (default \"info\") (debug, info, warn, error, dpanic, panic, fatal)")

	// Database - File
	databaseFilePath = flag.String("database-file-path", "./test/fixtures/file/metadata.json", "path to the database file")

	// HMAC
	hmacKey = flag.String("hmac-key", "", "hmac key to use for authentication between services")
)

func main() {
	ctx := context.Background()

	// Parse environment variables
	envy.Parse("PICSUM")

	// Parse commandline flags
	flag.Parse()

	// Initialize the logger
	log := logger.New(*loglevel)
	defer log.Sync()

	// Initialize tracing
	tracerCtx, tracerCancel := context.WithCancel(ctx)
	defer tracerCancel()

	tracer, err := tracing.New(tracerCtx, log, "picsum-photos")
	if err != nil {
		log.Fatalf("error initializing tracing: %s", err)
	}
	defer tracer.Shutdown(tracerCtx)

	// Set GOMAXPROCS
	maxprocs.Set(maxprocs.Logger(log.Infof))

	// Set up context for shutting down
	shutdownCtx, shutdown := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer shutdown()

	// Initialize the database
	database, err := fileDatabase.New(*databaseFilePath)
	if err != nil {
		log.Fatalf("error initializing database: %s", err)
	}

	// Initialize and start the health checker
	checkerCtx, checkerCancel := context.WithCancel(ctx)
	defer checkerCancel()

	checker := &health.Checker{
		Ctx:      checkerCtx,
		Database: database,
		Log:      log,
	}
	go checker.Run()

	// Start and listen on http
	api := &api.API{
		Database:        database,
		Log:             log,
		Tracer:          tracer,
		RootURL:         *rootURL,
		ImageServiceURL: *imageServiceURL,
		StaticPath:      staticPath,
		HandlerTimeout:  cmd.HandlerTimeout,
		HMAC: &hmac.HMAC{
			Key: []byte(*hmacKey),
		},
	}
	server := &http.Server{
		Addr:         *listen,
		Handler:      api.Router(),
		ReadTimeout:  cmd.ReadTimeout,
		WriteTimeout: cmd.WriteTimeout,
		ErrorLog:     logger.NewHTTPErrorLog(log),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Infof("error shutting down the http server: %s", err)
		}
	}()

	log.Infof("http server listening on %s", *listen)

	// Start the metrics http server
	go metrics.Serve(shutdownCtx, log, checker, *metricsListen)

	// Wait for shutdown
	<-shutdownCtx.Done()
	log.Infof("shutting down: %s", shutdownCtx.Err())

	// Shut down http server
	serverCtx, serverCancel := context.WithTimeout(ctx, cmd.WriteTimeout)
	defer serverCancel()
	if err := server.Shutdown(serverCtx); err != nil {
		log.Warnf("error shutting down: %s", err)
	}
}
