package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/DMarby/picsum-photos/internal/cache"
	"github.com/DMarby/picsum-photos/internal/cache/memory"
	"github.com/DMarby/picsum-photos/internal/cache/redis"
	"github.com/DMarby/picsum-photos/internal/cmd"
	"github.com/DMarby/picsum-photos/internal/health"
	"github.com/DMarby/picsum-photos/internal/hmac"
	"github.com/DMarby/picsum-photos/internal/image"
	"github.com/DMarby/picsum-photos/internal/image/vips"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/storage"
	fileStorage "github.com/DMarby/picsum-photos/internal/storage/file"
	"github.com/DMarby/picsum-photos/internal/storage/spaces"

	api "github.com/DMarby/picsum-photos/internal/imageapi"

	"github.com/jamiealquiza/envy"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

// Comandline flags
var (
	// Global
	listen   = flag.String("listen", ":8081", "listen address")
	loglevel = zap.LevelFlag("log-level", zap.InfoLevel, "log level (default \"info\") (debug, info, warn, error, dpanic, panic, fatal)")

	// Storage
	storageBackend = flag.String("storage", "file", "which storage backend to use (file, spaces)")

	// Storage - File
	storageFilePath = flag.String("storage-file-path", "./test/fixtures/file", "path to the file storage")

	// Storage - Spaces
	storageSpacesSpace     = flag.String("storage-spaces-space", "", "digitalocean space to use")
	storageSpacesRegion    = flag.String("storage-spaces-region", "", "spaces region")
	storageSpacesAccessKey = flag.String("storage-spaces-access-key", "", "spaces access key")
	storageSpacesSecretKey = flag.String("storage-spaces-secret-key", "", "spaces secret key")

	// Cache
	cacheBackend = flag.String("cache", "memory", "which cache backend to use (memory, redis)")

	// Cache - Redis
	cacheRedisAddress  = flag.String("cache-redis-address", "redis://127.0.0.1:6379", "redis address, may contain authentication details")
	cacheRedisPoolSize = flag.Int("cache-redis-pool-size", 10, "redis connection pool size")

	// Healthcheck
	healthCheckImageID = flag.String("health-check-image-id", "1", "image ID to request from the storage to check storage health")

	// HMAC
	hmacKey = flag.String("hmac-key", "", "hmac key to use for authentication between services")
)

func main() {
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
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	// Initialize the storage, cache
	storage, cache, err := setupBackends()
	if err != nil {
		log.Fatalf("error initializing backends: %s", err)
	}
	defer cache.Shutdown()

	// Initialize the image processor
	imageProcessorCtx, imageProcessorCancel := context.WithCancel(context.Background())
	defer imageProcessorCancel()

	imageProcessor, err := vips.New(imageProcessorCtx, log, image.NewCache(cache, storage))
	if err != nil {
		log.Fatalf("error initializing image processor %s", err.Error())
	}

	// Initialize and start the health checker
	checkerCtx, checkerCancel := context.WithCancel(context.Background())
	defer checkerCancel()

	checker := &health.Checker{
		Ctx:     checkerCtx,
		Storage: storage,
		ImageID: *healthCheckImageID,
		Cache:   cache,
		Log:     log,
	}
	go checker.Run()

	// Start and listen on http
	api := &api.API{
		ImageProcessor: imageProcessor,
		HealthChecker:  checker,
		Log:            log,
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
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Infof("shutting down the http server: %s", err)
			shutdown()
		}
	}()

	log.Infof("http server listening on %s", *listen)

	// Wait for shutdown or error
	err = cmd.WaitForInterrupt(shutdownCtx)
	log.Infof("shutting down: %s", err)

	// Shut down http server
	serverCtx, serverCancel := context.WithTimeout(context.Background(), cmd.WriteTimeout)
	defer serverCancel()
	if err := server.Shutdown(serverCtx); err != nil {
		log.Warnf("error shutting down: %s", err)
	}
}

func setupBackends() (storage storage.Provider, cache cache.Provider, err error) {
	// Storage
	switch *storageBackend {
	case "file":
		storage, err = fileStorage.New(*storageFilePath)
	case "spaces":
		storage, err = spaces.New(*storageSpacesSpace, *storageSpacesRegion, *storageSpacesAccessKey, *storageSpacesSecretKey)
	default:
		err = fmt.Errorf("invalid storage backend")
	}

	if err != nil {
		return
	}

	// Cache
	switch *cacheBackend {
	case "memory":
		cache = memory.New()
	case "redis":
		cache, err = redis.New(*cacheRedisAddress, *cacheRedisPoolSize)
	default:
		err = fmt.Errorf("invalid cache backend")
	}

	return
}
