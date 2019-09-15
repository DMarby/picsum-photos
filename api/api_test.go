package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"runtime"
	"time"

	"github.com/DMarby/picsum-photos/api"
	"github.com/DMarby/picsum-photos/health"
	"github.com/DMarby/picsum-photos/image"
	"github.com/DMarby/picsum-photos/logger"
	"go.uber.org/zap"

	"github.com/DMarby/picsum-photos/database"
	fileDatabase "github.com/DMarby/picsum-photos/database/file"
	mockDatabase "github.com/DMarby/picsum-photos/database/mock"

	mockProcessor "github.com/DMarby/picsum-photos/image/mock"
	vipsProcessor "github.com/DMarby/picsum-photos/image/vips"

	fileStorage "github.com/DMarby/picsum-photos/storage/file"
	mockStorage "github.com/DMarby/picsum-photos/storage/mock"

	memoryCache "github.com/DMarby/picsum-photos/cache/memory"
	mockCache "github.com/DMarby/picsum-photos/cache/mock"

	"testing"
)

const rootURL = "https://example.com"

func TestAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.New(zap.FatalLevel)
	defer log.Sync()

	storage, _ := fileStorage.New("../test/fixtures/file")
	db, _ := fileDatabase.New("../test/fixtures/file/metadata.json")
	dbMultiple, _ := fileDatabase.New("../test/fixtures/file/metadata_multiple.json")
	cache := memoryCache.New()
	imageCache := image.NewCache(cache, storage)
	imageProcessor, _ := vipsProcessor.New(ctx, log, imageCache)
	mockStorageImageProcessor, _ := vipsProcessor.New(ctx, log, image.NewCache(memoryCache.New(), &mockStorage.Provider{}))

	checker := &health.Checker{
		Ctx:      ctx,
		Storage:  storage,
		Database: db,
		Cache:    cache,
		Log:      log,
	}
	checker.Run()

	mockChecker := &health.Checker{
		Ctx:      ctx,
		Storage:  &mockStorage.Provider{},
		Database: &mockDatabase.Provider{},
		Cache:    &mockCache.Provider{},
		Log:      log,
	}
	mockChecker.Run()

	staticPath := "../src"

	router := (&api.API{imageProcessor, db, checker, log, 200, rootURL, staticPath, time.Minute}).Router()
	paginationRouter := (&api.API{imageProcessor, dbMultiple, checker, log, 200, rootURL, staticPath, time.Minute}).Router()
	mockStorageRouter := (&api.API{mockStorageImageProcessor, db, mockChecker, log, 200, rootURL, staticPath, time.Minute}).Router()
	mockProcessorRouter := (&api.API{&mockProcessor.Processor{}, db, checker, log, 200, rootURL, staticPath, time.Minute}).Router()
	mockDatabaseRouter := (&api.API{imageProcessor, &mockDatabase.Provider{}, checker, log, 200, rootURL, staticPath, time.Minute}).Router()

	tests := []struct {
		Name             string
		URL              string
		Router           http.Handler
		ExpectedStatus   int
		ExpectedResponse []byte
		ExpectedHeaders  map[string]string
	}{
		{
			Name:           "/v2/list lists images",
			URL:            "/v2/list",
			Router:         paginationRouter,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{
				api.ListImage{
					Image: database.Image{
						ID:     "1",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/1/300/400", rootURL),
				},
				api.ListImage{
					Image: database.Image{
						ID:     "2",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/2/300/400", rootURL),
				},
			}),
			ExpectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Link":          fmt.Sprintf("<%s/v2/list?page=2&limit=30>; rel=\"next\"", rootURL),
				"Cache-Control": "no-cache, no-store, must-revalidate",
			},
		},
		{
			Name:           "/v2/list lists images with limit",
			URL:            "/v2/list?limit=1000",
			Router:         paginationRouter,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{
				api.ListImage{
					Image: database.Image{
						ID:     "1",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/1/300/400", rootURL),
				},
				api.ListImage{
					Image: database.Image{
						ID:     "2",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/2/300/400", rootURL),
				},
			}),
			ExpectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Link":          fmt.Sprintf("<%s/v2/list?page=2&limit=100>; rel=\"next\"", rootURL),
				"Cache-Control": "no-cache, no-store, must-revalidate",
			},
		},
		{
			Name:           "/v2/list pagination page 1",
			URL:            "/v2/list?page=1&limit=1",
			Router:         paginationRouter,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{
				api.ListImage{
					Image: database.Image{
						ID:     "1",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/1/300/400", rootURL),
				},
			}),
			ExpectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Link":          fmt.Sprintf("<%s/v2/list?page=2&limit=1>; rel=\"next\"", rootURL),
				"Cache-Control": "no-cache, no-store, must-revalidate",
			},
		},
		{
			Name:           "/v2/list pagination page 2",
			URL:            "/v2/list?page=2&limit=1",
			Router:         paginationRouter,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{
				api.ListImage{
					Image: database.Image{
						ID:     "2",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/2/300/400", rootURL),
				},
			}),
			ExpectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Link":          fmt.Sprintf("<%s/v2/list?page=1&limit=1>; rel=\"prev\", <%s/v2/list?page=3&limit=1>; rel=\"next\"", rootURL, rootURL),
				"Cache-Control": "no-cache, no-store, must-revalidate",
			},
		},
		{
			Name:             "/v2/list pagination page 3",
			URL:              "/v2/list?page=3&limit=1",
			Router:           paginationRouter,
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{}),
			ExpectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Link":          fmt.Sprintf("<%s/v2/list?page=2&limit=1>; rel=\"prev\"", rootURL),
				"Cache-Control": "no-cache, no-store, must-revalidate",
			},
		},
		{
			Name:           "Deprecated /list lists images",
			URL:            "/list",
			Router:         router,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.DeprecatedImage{
				api.DeprecatedImage{
					Format:    "jpeg",
					Width:     300,
					Height:    400,
					Filename:  "1.jpeg",
					ID:        1,
					Author:    "John Doe",
					AuthorURL: "https://picsum.photos",
					PostURL:   "https://picsum.photos",
				},
			}),
			ExpectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Cache-Control": "no-cache, no-store, must-revalidate",
			},
		},
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
		{
			Name:           "/id/{id}/info returns info about an image",
			URL:            "/id/1/info",
			Router:         paginationRouter,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson(
				api.ListImage{
					Image: database.Image{
						ID:     "1",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/1/300/400", rootURL),
				},
			),
			ExpectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Cache-Control": "no-cache, no-store, must-revalidate",
			},
		},

		// Static page handling
		{"index", "/", router, http.StatusOK, readFile(path.Join(staticPath, "index.html")), map[string]string{"Content-Type": "text/html; charset=utf-8", "Cache-Control": "public, max-age=3600"}},
		{"images", "/images", router, http.StatusOK, readFile(path.Join(staticPath, "images.html")), map[string]string{"Content-Type": "text/html; charset=utf-8", "Cache-Control": "public, max-age=3600"}},
		{"favicon", "/assets/images/digitalocean.svg", router, http.StatusOK, readFile(path.Join(staticPath, "assets/images/digitalocean.svg")), map[string]string{"Content-Type": "image/svg+xml", "Cache-Control": "public, max-age=3600"}},

		// Errors
		{"invalid image id", "/id/nonexistant/200/300", router, http.StatusNotFound, []byte("Image does not exist\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid image id", "/id/nonexistant/info", router, http.StatusNotFound, []byte("Image does not exist\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid size", "/id/1/1/9223372036854775808", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}}, // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/9223372036854775808/1", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}}, // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/5500/1", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},                // Number larger then maxImageSize to fail int parsing
		{"invalid blur amount", "/id/1/100/100?blur=11", router, http.StatusBadRequest, []byte("Invalid blur amount\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid blur amount", "/id/1/100/100?blur=0", router, http.StatusBadRequest, []byte("Invalid blur amount\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid file extension", "/id/1/100/100.png", router, http.StatusBadRequest, []byte("Invalid file extension\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// Deprecated handler errors
		{"invalid size", "/g/9223372036854775808", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}}, // Number larger then max int size to fail int parsing
		// Storage errors
		{"Get() storage", "/id/1/100/100", mockStorageRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// Database errors
		{"List()", "/list", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"List()", "/v2/list", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"GetRandom()", "/200", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"GetRandomWithSeed()", "/seed/1/200", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"Get() database", "/id/1/100/100", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"Get() database info", "/id/1/info", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// Processor errors
		{"processor error", "/id/1/100/100", mockProcessorRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// 404
		{"404", "/asdf", router, http.StatusNotFound, []byte("page not found\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
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
		{"/id/:id/:width/:height", "/id/1/200/120", readFixture("width_height", "jpg"), "inline; filename=\"1-200x120.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg", "/id/1/200/120.jpg", readFixture("width_height", "jpg"), "inline; filename=\"1-200x120.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height?blur", "/id/1/200/200?blur", readFixture("blur", "jpg"), "inline; filename=\"1-200x200-blur_5.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?blur", "/id/1/200/200.jpg?blur", readFixture("blur", "jpg"), "inline; filename=\"1-200x200-blur_5.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height?grayscale", "/id/1/200/200?grayscale", readFixture("grayscale", "jpg"), "inline; filename=\"1-200x200-grayscale.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?grayscale", "/id/1/200/200.jpg?grayscale", readFixture("grayscale", "jpg"), "inline; filename=\"1-200x200-grayscale.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height?blur&grayscale", "/id/1/200/200?blur&grayscale", readFixture("all", "jpg"), "inline; filename=\"1-200x200-blur_5-grayscale.jpg\"", "image/jpeg"},
		{"/id/:id/:width/:height.jpg?blur&grayscale", "/id/1/200/200.jpg?blur&grayscale", readFixture("all", "jpg"), "inline; filename=\"1-200x200-blur_5-grayscale.jpg\"", "image/jpeg"},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400", readFixture("max_allowed", "jpg"), "inline; filename=\"1-300x400.jpg\"", "image/jpeg"},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400.jpg", readFixture("max_allowed", "jpg"), "inline; filename=\"1-300x400.jpg\"", "image/jpeg"},
		{"width/height of 0 returns original image width", "/id/1/0/0", readFixture("max_allowed", "jpg"), "inline; filename=\"1-300x400.jpg\"", "image/jpeg"},
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
		Name            string
		URL             string
		ExpectedURL     string
		TestCacheHeader bool
	}{
		// /id/:id/:size to /id/:id/:width/:height
		{"/id/:id/:size", "/id/1/200", "/id/1/200/200", true},
		{"/id/:id/:size.jpg", "/id/1/200.jpg", "/id/1/200/200.jpg", true},
		{"/id/:id/:size.webp", "/id/1/200.webp", "/id/1/200/200.webp", true},
		{"/id/:id/:size?blur", "/id/1/200?blur", "/id/1/200/200?blur=5", true},
		{"/id/:id/:size?grayscale", "/id/1/200?grayscale", "/id/1/200/200?grayscale", true},
		{"/id/:id/:size?blur&grayscale", "/id/1/200?blur&grayscale", "/id/1/200/200?blur=5&grayscale", true},
		// General
		{"/:size", "/200", "/id/1/200/200", true},
		{"/:width/:height", "/200/300", "/id/1/200/300", true},
		{"/:size.jpg", "/200.jpg", "/id/1/200/200.jpg", true},
		{"/:width/:height.jpg", "/200/300.jpg", "/id/1/200/300.jpg", true},
		{"/:size.webp", "/200.webp", "/id/1/200/200.webp", true},
		{"/:width/:height.webp", "/200/300.webp", "/id/1/200/300.webp", true},
		{"/:size?grayscale", "/200?grayscale", "/id/1/200/200?grayscale", true},
		{"/:width/:height?grayscale", "/200/300?grayscale", "/id/1/200/300?grayscale", true},
		// Default blur amount
		{"/:size?blur", "/200?blur", "/id/1/200/200?blur=5", true},
		{"/:width/:height?blur", "/200/300?blur", "/id/1/200/300?blur=5", true},
		{"/:size?grayscale&blur", "/200?grayscale&blur", "/id/1/200/200?blur=5&grayscale", true},
		{"/:width/:height?grayscale&blur", "/200/300?grayscale&blur", "/id/1/200/300?blur=5&grayscale", true},
		// Custom blur amount
		{"/:size?blur=10", "/200?blur=10", "/id/1/200/200?blur=10", true},
		{"/:width/:height?blur=10", "/200/300?blur=10", "/id/1/200/300?blur=10", true},
		{"/:size?grayscale&blur=10", "/200?grayscale&blur=10", "/id/1/200/200?blur=10&grayscale", true},
		{"/:width/:height?grayscale&blur=10", "/200/300?grayscale&blur=10", "/id/1/200/300?blur=10&grayscale", true},
		// Deprecated routes
		{"/g/:size", "/g/200", "/200/200?grayscale", true},
		{"/g/:width/:height", "/g/200/300", "/200/300?grayscale", true},
		{"/g/:size.jpg", "/g/200.jpg", "/200/200.jpg?grayscale", true},
		{"/g/:width/:height.jpg", "/g/200/300.jpg", "/200/300.jpg?grayscale", true},
		{"/g/:size.webp", "/g/200.webp", "/200/200.webp?grayscale", true},
		{"/g/:width/:height.webp", "/g/200/300.webp", "/200/300.webp?grayscale", true},
		{"/g/:size?blur", "/g/200?blur", "/200/200?blur=5&grayscale", true},
		{"/g/:width/:height?blur", "/g/200/300?blur", "/200/300?blur=5&grayscale", true},
		{"/g/:size?image=:id", "/g/200?image=1", "/id/1/200/200?grayscale", true},
		{"/g/:width/:height?image=:id", "/g/200/300?image=1", "/id/1/200/300?grayscale", true},
		{"/g/:size.jpg?image=:id", "/g/200.jpg?image=1", "/id/1/200/200.jpg?grayscale", true},
		{"/g/:width/:height.jpg?image=:id", "/g/200/300.jpg?image=1", "/id/1/200/300.jpg?grayscale", true},
		{"/g/:size.webp?image=:id", "/g/200.webp?image=1", "/id/1/200/200.webp?grayscale", true},
		{"/g/:width/:height.webp?image=:id", "/g/200/300.webp?image=1", "/id/1/200/300.webp?grayscale", true},
		// Deprecated query params
		{"/:size?image=:id", "/200?image=1", "/id/1/200/200", true},
		{"/:width/:height?image=:id", "/200/300?image=1", "/id/1/200/300", true},
		{"/:size?image=:id&grayscale", "/200?image=1&grayscale", "/id/1/200/200?grayscale", true},
		{"/:width/:height?image=:id&grayscale", "/200/300?image=1&grayscale", "/id/1/200/300?grayscale", true},
		{"/:size?image=:id&blur", "/200?image=1&blur", "/id/1/200/200?blur=5", true},
		{"/:width/:height?image=:id&blur", "/200/300?image=1&blur", "/id/1/200/300?blur=5", true},
		{"/:size?image=:id&grayscale&blur", "/200?image=1&grayscale&blur", "/id/1/200/200?blur=5&grayscale", true},
		{"/:width/:height?image=:id&grayscale&blur", "/200/300?image=1&grayscale&blur", "/id/1/200/300?blur=5&grayscale", true},
		// By seed
		{"/seed/:seed/:size", "/seed/1/200", "/id/1/200/200", true},
		{"/seed/:seed/:size.jpg", "/seed/1/200.jpg", "/id/1/200/200.jpg", true},
		{"/seed/:seed/:size.webp", "/seed/1/200.webp", "/id/1/200/200.webp", true},
		{"/seed/:seed/:size?blur", "/seed/1/200?blur", "/id/1/200/200?blur=5", true},
		{"/seed/:seed/:size?blur=10", "/seed/1/200?blur=10", "/id/1/200/200?blur=10", true},
		{"/seed/:seed/:size?grayscale", "/seed/1/200?grayscale", "/id/1/200/200?grayscale", true},
		{"/seed/:seed/:size?blur&grayscale", "/seed/1/200?blur&grayscale", "/id/1/200/200?blur=5&grayscale", true},
		{"/seed/:seed/:size?blur=10&grayscale", "/seed/1/200?blur=10&grayscale", "/id/1/200/200?blur=10&grayscale", true},
		{"/seed/:seed/:width/:height", "/seed/1/200/300", "/id/1/200/300", true},
		{"/seed/:seed/:width/:height.jpg", "/seed/1/200/300.jpg", "/id/1/200/300.jpg", true},
		{"/seed/:seed/:width/:height.webp", "/seed/1/200/300.webp", "/id/1/200/300.webp", true},
		{"/seed/:seed/:width/:height?blur", "/seed/1/200/300?blur", "/id/1/200/300?blur=5", true},
		{"/seed/:seed/:width/:height?blur=10", "/seed/1/200/300?blur=10", "/id/1/200/300?blur=10", true},
		{"/seed/:seed/:width/:height?grayscale", "/seed/1/200/300?grayscale", "/id/1/200/300?grayscale", true},
		{"/seed/:seed/:width/:height?blur&grayscale", "/seed/1/200/300?blur&grayscale", "/id/1/200/300?blur=5&grayscale", true},
		{"/seed/:seed/:width/:height?blur=10&grayscale", "/seed/1/200/300?blur=10&grayscale", "/id/1/200/300?blur=10&grayscale", true},
		// Trailing slashes
		{"/:size/", "/200/", "/200", false},
		{"/:width/:height/", "/200/300/", "/200/300", false},
		{"/id/:id/:size/", "/id/1/200/", "/id/1/200", false},
		{"/id/:id/:width/:height/", "/id/1/200/120/", "/id/1/200/120", false},
		{"/seed/:seed/:size/", "/seed/1/200/", "/seed/1/200", false},
		{"/seed/:seed/:width/:height/", "/seed/1/200/120/", "/seed/1/200/120", false},
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

		if test.TestCacheHeader {
			if cacheControl := w.Header().Get("Cache-Control"); cacheControl != "no-cache, no-store, must-revalidate" {
				t.Errorf("%s: wrong cache header, %#v", test.Name, cacheControl)
			}
		}
	}
}

func marshalJson(v interface{}) []byte {
	fixture, _ := json.Marshal(v)
	return append(fixture[:], []byte("\n")...)
}

func readFixture(fixtureName string, extension string) []byte {
	return readFile(fmt.Sprintf("../test/fixtures/api/%s_%s.%s", fixtureName, runtime.GOOS, extension))
}
func readFile(path string) []byte {
	fixture, _ := ioutil.ReadFile(path)
	return fixture
}
