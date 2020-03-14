package imageapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/health"
	"github.com/DMarby/picsum-photos/internal/image"
	api "github.com/DMarby/picsum-photos/internal/imageapi"
	"github.com/DMarby/picsum-photos/internal/logger"
	"go.uber.org/zap"

	fileDatabase "github.com/DMarby/picsum-photos/internal/database/file"
	mockDatabase "github.com/DMarby/picsum-photos/internal/database/mock"

	mockProcessor "github.com/DMarby/picsum-photos/internal/image/mock"
	vipsProcessor "github.com/DMarby/picsum-photos/internal/image/vips"

	fileStorage "github.com/DMarby/picsum-photos/internal/storage/file"
	mockStorage "github.com/DMarby/picsum-photos/internal/storage/mock"

	memoryCache "github.com/DMarby/picsum-photos/internal/cache/memory"
	mockCache "github.com/DMarby/picsum-photos/internal/cache/mock"

	"testing"
)

const rootURL = "https://example.com"

func TestAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log, db, imageProcessor, checker := setup(t, ctx)
	defer log.Sync()

	mockStorageImageProcessor, _ := vipsProcessor.New(ctx, log, image.NewCache(memoryCache.New(), &mockStorage.Provider{}))

	mockChecker := &health.Checker{
		Ctx:      ctx,
		Storage:  &mockStorage.Provider{},
		Database: &mockDatabase.Provider{},
		Cache:    &mockCache.Provider{},
		Log:      log,
	}
	mockChecker.Run()

	router := (&api.API{imageProcessor, db, checker, log, time.Minute}).Router()
	mockStorageRouter := (&api.API{mockStorageImageProcessor, db, mockChecker, log, time.Minute}).Router()
	mockProcessorRouter := (&api.API{&mockProcessor.Processor{}, db, checker, log, time.Minute}).Router()
	mockDatabaseRouter := (&api.API{imageProcessor, &mockDatabase.Provider{}, checker, log, time.Minute}).Router()

	tests := []struct {
		Name             string
		URL              string
		Router           http.Handler
		ExpectedStatus   int
		ExpectedResponse []byte
		ExpectedHeaders  map[string]string
	}{
		{
			Name:           "/health returns healthy health status",
			URL:            "/health",
			Router:         router,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson(health.Status{
				Healthy:  true,
				Cache:    "healthy",
				Database: "healthy",
				Storage:  "healthy",
			}),
			ExpectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			Name:           "/health returns unhealthy health status",
			URL:            "/health",
			Router:         mockStorageRouter,
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: marshalJson(health.Status{
				Healthy:  false,
				Cache:    "unhealthy",
				Database: "unhealthy",
				Storage:  "unknown",
			}),
			ExpectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},

		// Errors
		{"invalid image id", "/id/nonexistant/200/300.jpg", router, http.StatusNotFound, []byte("Image does not exist\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid size", "/id/1/1/9223372036854775808.jpg", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}}, // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/9223372036854775808/1.jpg", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}}, // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/5500/1.jpg", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},                // Number larger then maxImageSize to fail int parsing
		{"invalid blur amount", "/id/1/100/100.jpg?blur=11", router, http.StatusBadRequest, []byte("Invalid blur amount\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid blur amount", "/id/1/100/100.jpg?blur=0", router, http.StatusBadRequest, []byte("Invalid blur amount\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid file extension", "/id/1/100/100.png", router, http.StatusBadRequest, []byte("Invalid file extension\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// Storage errors
		{"Get() storage", "/id/1/100/100.jpg", mockStorageRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// Database errors
		{"Get() database", "/id/1/100/100.jpg", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// 404
		{"404", "/asdf", router, http.StatusNotFound, []byte("page not found\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// Processor errors
		{"processor error", "/id/1/100/100.jpg", mockProcessorRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", test.URL, nil)
		test.Router.ServeHTTP(w, req)
		if w.Code != test.ExpectedStatus {
			t.Errorf("%s: wrong response code, %#v", test.Name, w.Code)
			continue
		}

		if test.ExpectedHeaders != nil {
			for expectedHeader, expectedValue := range test.ExpectedHeaders {
				headerValue := w.Header().Get(expectedHeader)
				if headerValue != expectedValue {
					t.Errorf("%s: wrong header value for %s, %#v", test.Name, expectedHeader, headerValue)
				}
			}
		}

		if !reflect.DeepEqual(w.Body.Bytes(), test.ExpectedResponse) {
			t.Errorf("%s: wrong response %#v", test.Name, w.Body.String())
		}
	}

	imageTests := []struct {
		Name                       string
		URL                        string
		ExpectedResponse           []byte
		ExpectedContentDisposition string
		ExpectedContentType        string
	}{
		// Images

		// JPEG
		{"/id/:id/:width/:height.jpg", "/id/1/200/120.jpg", readFixture("width_height", "jpg"), "inline; filename=\"1-200x120.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?blur", "/id/1/200/200.jpg?blur", readFixture("blur", "jpg"), "inline; filename=\"1-200x200-blur_5.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?grayscale", "/id/1/200/200.jpg?grayscale", readFixture("grayscale", "jpg"), "inline; filename=\"1-200x200-grayscale.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?blur&grayscale", "/id/1/200/200.jpg?blur&grayscale", readFixture("all", "jpg"), "inline; filename=\"1-200x200-blur_5-grayscale.jpg\"", "image/jpeg"},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400.jpg", readFixture("max_allowed", "jpg"), "inline; filename=\"1-300x400.jpg\"", "image/jpeg"},
		{"width/height of 0 returns original image width", "/id/1/0/0.jpg", readFixture("max_allowed", "jpg"), "inline; filename=\"1-300x400.jpg\"", "image/jpeg"},

		// WebP
		{"/id/:id/:width/:height.webp", "/id/1/200/120.webp", readFixture("width_height", "webp"), "inline; filename=\"1-200x120.webp\"", "image/webp"},
		{"/id/:id/:width/:height.webp?blur", "/id/1/200/200.webp?blur", readFixture("blur", "webp"), "inline; filename=\"1-200x200-blur_5.webp\"", "image/webp"},
		{"/id/:id/:width/:height.webp?grayscale", "/id/1/200/200.webp?grayscale", readFixture("grayscale", "webp"), "inline; filename=\"1-200x200-grayscale.webp\"", "image/webp"},
		{"/id/:id/:width/:height.webp?blur&grayscale", "/id/1/200/200.webp?blur&grayscale", readFixture("all", "webp"), "inline; filename=\"1-200x200-blur_5-grayscale.webp\"", "image/webp"},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400.webp", readFixture("max_allowed", "webp"), "inline; filename=\"1-300x400.webp\"", "image/webp"},
		{"width/height of 0 returns original image width", "/id/1/0/0.webp", readFixture("max_allowed", "webp"), "inline; filename=\"1-300x400.webp\"", "image/webp"},
	}

	for _, test := range imageTests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", test.URL, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s: wrong response code, %#v", test.Name, w.Code)
			continue
		}

		if contentType := w.Header().Get("Content-Type"); contentType != test.ExpectedContentType {
			t.Errorf("%s: wrong content type, %#v", test.Name, contentType)
		}

		if cacheControl := w.Header().Get("Cache-Control"); cacheControl != "public, max-age=2592000" {
			t.Errorf("%s: wrong cache header, %#v", test.Name, cacheControl)
		}

		if contentDisposition := w.Header().Get("Content-Disposition"); contentDisposition != test.ExpectedContentDisposition {
			t.Errorf("%s: wrong content disposition header, %#v", test.Name, contentDisposition)
		}

		if imageID := w.Header().Get("Picsum-ID"); imageID != "1" {
			t.Errorf("%s: wrong image id header, %#v", test.Name, imageID)
		}

		if !reflect.DeepEqual(w.Body.Bytes(), test.ExpectedResponse) {
			t.Errorf("%s: wrong response/image data", test.Name)
		}
	}

	redirectTests := []struct {
		Name        string
		URL         string
		ExpectedURL string
	}{
		// Trailing slashes
		{"/id/:id/:width/:height/", "/id/1/200/120.jpg/", "/id/1/200/120.jpg"},
	}

	for _, test := range redirectTests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", test.URL, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusFound && w.Code != http.StatusMovedPermanently {
			t.Errorf("%s: wrong response code, %#v", test.Name, w.Code)
			continue
		}

		location := w.Header().Get("Location")
		if location != test.ExpectedURL {
			t.Errorf("%s: wrong redirect %s", test.Name, location)
		}
	}
}

func marshalJson(v interface{}) []byte {
	fixture, _ := json.Marshal(v)
	return append(fixture[:], []byte("\n")...)
}

func readFixture(fixtureName string, extension string) []byte {
	return readFile(fixturePath(fixtureName, extension))
}

// Utility function for regenerating the fixtures
func TestFixtures(t *testing.T) {
	if os.Getenv("GENERATE_FIXTURES") != "1" {
		t.SkipNow()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log, db, imageProcessor, checker := setup(t, ctx)
	defer log.Sync()

	router := (&api.API{imageProcessor, db, checker, log, time.Minute}).Router()

	// JPEG
	createFixture(router, "/id/1/200/120.jpg", "width_height", "jpg")
	createFixture(router, "/id/1/200/200.jpg?blur", "blur", "jpg")
	createFixture(router, "/id/1/200/200.jpg?grayscale", "grayscale", "jpg")
	createFixture(router, "/id/1/200/200.jpg?blur&grayscale", "all", "jpg")
	createFixture(router, "/id/1/300/400.jpg", "max_allowed", "jpg")

	// WebP
	createFixture(router, "/id/1/200/120.webp", "width_height", "webp")
	createFixture(router, "/id/1/200/200.webp?blur", "blur", "webp")
	createFixture(router, "/id/1/200/200.webp?grayscale", "grayscale", "webp")
	createFixture(router, "/id/1/200/200.webp?blur&grayscale", "all", "webp")
	createFixture(router, "/id/1/300/400.webp", "max_allowed", "webp")
}

func setup(t *testing.T, ctx context.Context) (*logger.Logger, database.Provider, image.Processor, *health.Checker) {
	t.Helper()

	log := logger.New(zap.FatalLevel)

	storage, _ := fileStorage.New("../../test/fixtures/file")
	db, _ := fileDatabase.New("../../test/fixtures/file/metadata.json")
	cache := memoryCache.New()
	imageCache := image.NewCache(cache, storage)
	imageProcessor, _ := vipsProcessor.New(ctx, log, imageCache)

	checker := &health.Checker{
		Ctx:      ctx,
		Storage:  storage,
		Database: db,
		Cache:    cache,
		Log:      log,
	}
	checker.Run()

	return log, db, imageProcessor, checker
}

func createFixture(router http.Handler, URL string, fixtureName string, extension string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", URL, nil)
	router.ServeHTTP(w, req)
	ioutil.WriteFile(fixturePath(fixtureName, extension), w.Body.Bytes(), 644)
}

func fixturePath(fixtureName string, extension string) string {
	return fmt.Sprintf("../../test/fixtures/api/%s_%s.%s", fixtureName, runtime.GOOS, extension)
}

func readFile(path string) []byte {
	fixture, _ := ioutil.ReadFile(path)
	return fixture
}
