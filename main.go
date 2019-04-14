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
	"github.com/DMarby/picsum-photos/cache"
	"github.com/DMarby/picsum-photos/cache/memory"
	"github.com/DMarby/picsum-photos/cache/redis"
	"github.com/DMarby/picsum-photos/database"
	fileDatabase "github.com/DMarby/picsum-photos/database/file"
	"github.com/DMarby/picsum-photos/database/postgresql"
	"github.com/DMarby/picsum-photos/health"
	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/image/vips"
	"github.com/DMarby/picsum-photos/logger"
	"github.com/DMarby/picsum-photos/storage"
	fileStorage "github.com/DMarby/picsum-photos/storage/file"
	"github.com/DMarby/picsum-photos/storage/spaces"

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

// Http timeouts
const (
	readTimeout    = 5 * time.Second
	writeTimeout   = time.Minute
	handlerTimeout = 45 * time.Second
)

const (
	maxImageSize = 5000       // The max allowed image width/height to be requested
	staticPath   = "./static" // Path where the static files are located
)

// Comandline flags
var (
	// Global
	listen   = flag.String("listen", ":8080", "listen address")
	rootURL  = flag.String("root-url", "https://picsum.photos", "root url")
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

	// Database
	databaseBackend = flag.String("database", "file", "which database backend to use (file, postgresql)")

	// Database - File
	databaseFilePath = flag.String("database-file-path", "./test/fixtures/file/metadata.json", "path to the database file")

	// Database - Postgresql
	databasePostgresqlAddress = flag.String("database-postgresql-address", "postgresql://postgres@127.0.0.1/postgres", "postgresql address")
)

func main() {
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

	// Initialize the storage, cache and database
	storage, cache, database, err := setupBackends()
	if err != nil {
		log.Fatalf("error initializing backends: %s", err)
	}
	defer cache.Shutdown()
	defer database.Shutdown()

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
		Ctx:      checkerCtx,
		Storage:  storage,
		Database: database,
		Cache:    cache,
	}
	go checker.Run()

	// Start and listen on http
	api := &api.API{
		ImageProcessor: imageProcessor,
		Database:       database,
		HealthChecker:  checker,
		Log:            log,
		MaxImageSize:   maxImageSize,
		RootURL:        *rootURL,
		StaticPath:     staticPath,
		HandlerTimeout: handlerTimeout,
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

func setupBackends() (storage storage.Provider, cache cache.Provider, database database.Provider, err error) {
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

	if err != nil {
		return
	}

	// Database
	switch *databaseBackend {
	case "file":
		database, err = fileDatabase.New(*databaseFilePath)
	case "postgresql":
		database, err = postgresql.New(*databasePostgresqlAddress)
	default:
		err = fmt.Errorf("invalid database backend")
	}

	return
}
