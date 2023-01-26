package imageapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/DMarby/picsum-photos/internal/hmac"
	"github.com/DMarby/picsum-photos/internal/image"
	api "github.com/DMarby/picsum-photos/internal/imageapi"
	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/params"
	"github.com/DMarby/picsum-photos/internal/tracing"
	"github.com/DMarby/picsum-photos/internal/tracing/test"
	"go.uber.org/zap"

	mockProcessor "github.com/DMarby/picsum-photos/internal/image/mock"
	vipsProcessor "github.com/DMarby/picsum-photos/internal/image/vips"

	fileStorage "github.com/DMarby/picsum-photos/internal/storage/file"
	mockStorage "github.com/DMarby/picsum-photos/internal/storage/mock"

	memoryCache "github.com/DMarby/picsum-photos/internal/cache/memory"

	"testing"
)

func TestAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log, tracer, imageProcessor, hmac := setup(t, ctx)

	mockStorageImageProcessor, _ := vipsProcessor.New(ctx, log, tracer, 3, image.NewCache(tracer, memoryCache.New(), &mockStorage.Provider{}))

	router := (&api.API{imageProcessor, log, tracer, time.Minute, hmac}).Router()
	mockStorageRouter := (&api.API{mockStorageImageProcessor, log, tracer, time.Minute, hmac}).Router()
	mockProcessorRouter := (&api.API{&mockProcessor.Processor{}, log, tracer, time.Minute, hmac}).Router()

	tests := []struct {
		Name             string
		URL              string
		Router           http.Handler
		ExpectedStatus   int
		ExpectedResponse []byte
		ExpectedHeaders  map[string]string
		HMAC             bool
	}{
		// Errors
		{"invalid parameters", "/id/nonexistant/200/300.jpg", router, http.StatusBadRequest, []byte("Invalid parameters\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "private, no-cache, no-store, must-revalidate"}, false},
		// Storage errors
		{"Get() storage", "/id/1/100/100.jpg", mockStorageRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "private, no-cache, no-store, must-revalidate"}, true},
		// 404
		{"404", "/asdf", router, http.StatusNotFound, []byte("page not found\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "private, no-cache, no-store, must-revalidate"}, true},
		// Processor errors
		{"processor error", "/id/1/100/100.jpg", mockProcessorRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "private, no-cache, no-store, must-revalidate"}, true},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		if test.HMAC {
			url, err := params.HMAC(hmac, test.URL, url.Values{})
			if err != nil {
				t.Errorf("%s: hmac error %s", test.Name, err)
				continue
			}

			test.URL = url
		}

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
		{"/id/:id/:width/:height.jpg?blur=5", "/id/1/200/200.jpg?blur=5", readFixture("blur", "jpg"), "inline; filename=\"1-200x200-blur_5.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?grayscale", "/id/1/200/200.jpg?grayscale", readFixture("grayscale", "jpg"), "inline; filename=\"1-200x200-grayscale.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?blur=5&grayscale", "/id/1/200/200.jpg?blur=5&grayscale", readFixture("all", "jpg"), "inline; filename=\"1-200x200-blur_5-grayscale.jpg\"", "image/jpeg"},

		// WebP
		{"/id/:id/:width/:height.webp", "/id/1/200/120.webp", readFixture("width_height", "webp"), "inline; filename=\"1-200x120.webp\"", "image/webp"},
		{"/id/:id/:width/:height.webp?blur=5", "/id/1/200/200.webp?blur=5", readFixture("blur", "webp"), "inline; filename=\"1-200x200-blur_5.webp\"", "image/webp"},
		{"/id/:id/:width/:height.webp?grayscale", "/id/1/200/200.webp?grayscale", readFixture("grayscale", "webp"), "inline; filename=\"1-200x200-grayscale.webp\"", "image/webp"},
		{"/id/:id/:width/:height.webp?blur=5&grayscale", "/id/1/200/200.webp?blur=5&grayscale", readFixture("all", "webp"), "inline; filename=\"1-200x200-blur_5-grayscale.webp\"", "image/webp"},
	}

	for _, test := range imageTests {
		w := httptest.NewRecorder()

		u, err := url.Parse(test.URL)
		if err != nil {
			t.Errorf("%s: url error %s", test.Name, err)
			continue
		}

		url, err := params.HMAC(hmac, u.Path, u.Query())
		if err != nil {
			t.Errorf("%s: hmac error %s", test.Name, err)
			continue
		}

		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s: wrong response code, %#v", test.Name, w.Code)
			continue
		}

		if contentType := w.Header().Get("Content-Type"); contentType != test.ExpectedContentType {
			t.Errorf("%s: wrong content type, %#v", test.Name, contentType)
		}

		if cacheControl := w.Header().Get("Cache-Control"); cacheControl != "public, max-age=2592000, stale-while-revalidate=60, stale-if-error=43200, immutable" {
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

	log, tracer, imageProcessor, hmac := setup(t, ctx)

	router := (&api.API{imageProcessor, log, tracer, time.Minute, hmac}).Router()

	// JPEG
	createFixture(router, hmac, "/id/1/200/120.jpg", "width_height", "jpg")
	createFixture(router, hmac, "/id/1/200/200.jpg?blur=5", "blur", "jpg")
	createFixture(router, hmac, "/id/1/200/200.jpg?grayscale", "grayscale", "jpg")
	createFixture(router, hmac, "/id/1/200/200.jpg?blur=5&grayscale", "all", "jpg")
	createFixture(router, hmac, "/id/1/300/400.jpg", "max_allowed", "jpg")

	// WebP
	createFixture(router, hmac, "/id/1/200/120.webp", "width_height", "webp")
	createFixture(router, hmac, "/id/1/200/200.webp?blur=5", "blur", "webp")
	createFixture(router, hmac, "/id/1/200/200.webp?grayscale", "grayscale", "webp")
	createFixture(router, hmac, "/id/1/200/200.webp?blur=5&grayscale", "all", "webp")
	createFixture(router, hmac, "/id/1/300/400.webp", "max_allowed", "webp")
}

func setup(t *testing.T, ctx context.Context) (*logger.Logger, *tracing.Tracer, image.Processor, *hmac.HMAC) {
	t.Helper()

	log := logger.New(zap.FatalLevel)
	tracer := test.Tracer(log)

	storage, _ := fileStorage.New("../../test/fixtures/file")
	cache := memoryCache.New()
	imageCache := image.NewCache(tracer, cache, storage)
	imageProcessor, _ := vipsProcessor.New(ctx, log, tracer, 3, imageCache)

	hmac := &hmac.HMAC{
		Key: []byte("test"),
	}

	t.Cleanup(func() {
		log.Sync()
	})

	return log, tracer, imageProcessor, hmac
}

func createFixture(router http.Handler, hmac *hmac.HMAC, URL string, fixtureName string, extension string) {
	w := httptest.NewRecorder()

	u, _ := url.Parse(URL)
	url, _ := params.HMAC(hmac, u.Path, u.Query())

	req, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)
	os.WriteFile(fixturePath(fixtureName, extension), w.Body.Bytes(), 0644)
}

func fixturePath(fixtureName string, extension string) string {
	return fmt.Sprintf("../../test/fixtures/api/%s_%s.%s", fixtureName, runtime.GOOS, extension)
}

func readFile(path string) []byte {
	fixture, _ := os.ReadFile(path)
	return fixture
}
