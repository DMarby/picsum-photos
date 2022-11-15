package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DMarby/picsum-photos/internal/cache/redis"
	"github.com/DMarby/picsum-photos/internal/cmd"
	"github.com/DMarby/picsum-photos/internal/health"
	"github.com/DMarby/picsum-photos/internal/hmac"
	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/image/vips"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/metrics"
	"github.com/DMarby/picsum-photos/internal/storage/spaces"
	"github.com/DMarby/picsum-photos/internal/tracing"

	api "github.com/DMarby/picsum-photos/internal/imageapi"

	"github.com/jamiealquiza/envy"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

// Comandline flags
var (
	// Global
	listen        = flag.String("listen", ":8081", "listen address")
	metricsListen = flag.String("metrics-listen", ":8083", "metrics listen address")
	loglevel      = zap.LevelFlag("log-level", zap.InfoLevel, "log level (default \"info\") (debug, info, warn, error, dpanic, panic, fatal)")

	// Storage - Spaces
	storageSpacesSpace          = flag.String("storage-spaces-space", "", "digitalocean space to use")
	storageSpacesEndpoint       = flag.String("storage-spaces-endpoint", "", "spaces endpoint")
	storageSpacesForcePathStyle = flag.Bool("storage-spaces-force-path-style", false, "spaces force path style")
	storageSpacesAccessKey      = flag.String("storage-spaces-access-key", "", "spaces access key")
	storageSpacesSecretKey      = flag.String("storage-spaces-secret-key", "", "spaces secret key")

	// Cache - Redis
	cacheRedisAddress  = flag.String("cache-redis-address", "redis://127.0.0.1:6379", "redis address, may contain authentication details")
	cacheRedisPoolSize = flag.Int("cache-redis-pool-size", 10, "redis connection pool size")

	// HMAC
	hmacKey = flag.String("hmac-key", "", "hmac key to use for authentication between services")

	// Image processor
	workers = flag.Int("workers", 3, "worker queue concurrency")
)

func main() {
	ctx := context.Background()

	// Parse environment variables
	envy.Parse("IMAGE")

	// Parse commandline flags
	flag.Parse()

	// Initialize the logger
	log := logger.New(*loglevel)
	defer log.Sync()

	// Set GOMAXPROCS
	maxprocs.Set(maxprocs.Logger(log.Infof))

	// Set up context for shutting down
	shutdownCtx, shutdown := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer shutdown()

	// Initialize tracing
	tracerCtx, tracerCancel := context.WithCancel(ctx)
	defer tracerCancel()

	tracer, err := tracing.New(tracerCtx, log, "image-service")
	if err != nil {
		log.Fatalf("error initializing tracing: %s", err)
	}
	defer tracer.Shutdown(tracerCtx)

	// Initialize the storage, cache
	storage, err := spaces.New(*storageSpacesSpace, *storageSpacesEndpoint, *storageSpacesAccessKey, *storageSpacesSecretKey, *storageSpacesForcePathStyle)
	if err != nil {
		log.Fatalf("error initializing storage: %s", err)
	}

	// Initialize the cache
	cache, err := redis.New(ctx, tracer, *cacheRedisAddress, *cacheRedisPoolSize)
	if err != nil {
		log.Fatalf("error initializing cache: %s", err)
	}
	defer cache.Shutdown()

	// Initialize the image processor
	imageProcessor, err := vips.New(shutdownCtx, log, tracer, *workers, image.NewCache(tracer, cache, storage))
	if err != nil {
		log.Fatalf("error initializing image processor %s", err.Error())
	}

	// Initialize and start the health checker
	checker := &health.Checker{
		Ctx:     shutdownCtx,
		Storage: storage,
		Cache:   cache,
		Log:     log,
	}
	go checker.Run()

	// Start and listen on http
	api := &api.API{
		ImageProcessor: imageProcessor,
		Log:            log,
		Tracer:         tracer,
		HandlerTimeout: cmd.HandlerTimeout,
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
			log.Errorf("error shutting down the http server: %s", err)
		}
	}()

	log.Infof("http server listening on %s", *listen)

	// Start the metrics http server
	go metrics.Serve(shutdownCtx, log, checker, *metricsListen)

	// Wait for shutdown
	<-shutdownCtx.Done()
	log.Infof("shutting down: %s", shutdownCtx.Err())

	// Shut down http server
	serverCtx, serverCancel := context.WithTimeout(context.Background(), cmd.WriteTimeout)
	defer serverCancel()
	if err := server.Shutdown(serverCtx); err != nil {
		log.Warnf("error shutting down: %s", err)
	}
}
