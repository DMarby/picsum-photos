package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"

	"github.com/DMarby/picsum-photos/api"
	"github.com/DMarby/picsum-photos/health"
	"github.com/DMarby/picsum-photos/logger"
	"go.uber.org/zap"

	"github.com/DMarby/picsum-photos/database"
	fileDatabase "github.com/DMarby/picsum-photos/database/file"
	mockDatabase "github.com/DMarby/picsum-photos/database/mock"

	mockProcessor "github.com/DMarby/picsum-photos/image/mock"
	vipsProcessor "github.com/DMarby/picsum-photos/image/vips"

	fileStorage "github.com/DMarby/picsum-photos/storage/file"
	mockStorage "github.com/DMarby/picsum-photos/storage/mock"

	"testing"
)

func marshalJson(v interface{}) []byte {
	fixture, _ := json.Marshal(v)
	return append(fixture[:], []byte("\n")...)
}

func readFixture(fixtureName string) []byte {
	fixture, _ := ioutil.ReadFile(fmt.Sprintf("../test/fixtures/api/%s_%s.jpg", fixtureName, runtime.GOOS))
	return fixture
}

const rootURL = "https://example.com"

func TestAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.New(zap.ErrorLevel)
	defer log.Sync()

	imageProcessor, _ := vipsProcessor.GetInstance(ctx, log)
	storage, _ := fileStorage.New("../test/fixtures/file")
	db, _ := fileDatabase.New("../test/fixtures/file/metadata.json")

	checker := health.New(ctx, imageProcessor, storage, db)
	checker.Run()

	mockChecker := health.New(ctx, &mockProcessor.Processor{}, &mockStorage.Provider{}, &mockDatabase.Provider{})
	mockChecker.Run()

	router := (&api.API{imageProcessor, storage, db, checker, log, 200, rootURL}).Router()
	mockStorageRouter := (&api.API{imageProcessor, &mockStorage.Provider{}, db, mockChecker, log, 200, rootURL}).Router()
	mockProcessorRouter := (&api.API{&mockProcessor.Processor{}, storage, db, checker, log, 200, rootURL}).Router()
	mockDatabaseRouter := (&api.API{imageProcessor, storage, &mockDatabase.Provider{}, checker, log, 200, rootURL}).Router()

	tests := []struct {
		Name                string
		URL                 string
		Router              http.Handler
		ExpectedStatus      int
		ExpectedResponse    []byte
		ExpectedContentType string
	}{
		{
			Name:           "/v2/list lists images",
			URL:            "/v2/list",
			Router:         router,
			ExpectedStatus: 200,
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
			ExpectedContentType: "application/json",
		},
		{
			Name:           "Deprecated /list lists images",
			URL:            "/list",
			Router:         router,
			ExpectedStatus: 200,
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
			ExpectedContentType: "application/json",
		},
		{
			Name:           "/health returns health status",
			URL:            "/health",
			Router:         router,
			ExpectedStatus: 200,
			ExpectedResponse: marshalJson(health.Status{
				Healthy:   true,
				Database:  "healthy",
				Storage:   "healthy",
				Processor: "healthy",
			}),
			ExpectedContentType: "application/json",
		},
		{
			Name:           "/health returns health status",
			URL:            "/health",
			Router:         mockStorageRouter,
			ExpectedStatus: 500,
			ExpectedResponse: marshalJson(health.Status{
				Healthy:   false,
				Database:  "unhealthy",
				Storage:   "unknown",
				Processor: "unknown",
			}),
			ExpectedContentType: "application/json",
		},
		// Errors
		{"invalid image id", "/id/2/200", router, 404, []byte("Image does not exist\n"), "text/plain; charset=utf-8"},
		{"invalid image id", "/id/2/200/300", router, 404, []byte("Image does not exist\n"), "text/plain; charset=utf-8"},
		{"invalid size", "/id/1/0", router, 400, []byte("Invalid size\n"), "text/plain; charset=utf-8"},
		{"invalid size", "/id/1/0/0", router, 400, []byte("Invalid size\n"), "text/plain; charset=utf-8"},
		{"invalid size", "/id/1/1/9223372036854775808", router, 400, []byte("Invalid size\n"), "text/plain; charset=utf-8"}, // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/9223372036854775808/1", router, 400, []byte("Invalid size\n"), "text/plain; charset=utf-8"}, // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/5500/1", router, 400, []byte("Invalid size\n"), "text/plain; charset=utf-8"},                // Number larger then maxImageSize to fail int parsing
		{"invalid blur amount", "/id/1/100/100?blur=11", router, 400, []byte("Invalid blur amount\n"), "text/plain; charset=utf-8"},
		{"invalid blur amount", "/id/1/100/100?blur=0", router, 400, []byte("Invalid blur amount\n"), "text/plain; charset=utf-8"},
		// Storage errors
		{"Get()", "/id/1/100", mockStorageRouter, 500, []byte("Something went wrong\n"), "text/plain; charset=utf-8"},
		// Database errors
		{"List()", "/list", mockDatabaseRouter, 500, []byte("Something went wrong\n"), "text/plain; charset=utf-8"},
		{"List()", "/v2/list", mockDatabaseRouter, 500, []byte("Something went wrong\n"), "text/plain; charset=utf-8"},
		{"GetRandom()", "/200", mockDatabaseRouter, 500, []byte("Something went wrong\n"), "text/plain; charset=utf-8"},
		{"Get()", "/id/1/100", mockDatabaseRouter, 500, []byte("Something went wrong\n"), "text/plain; charset=utf-8"},
		// Processor errors
		{"processor error", "/id/1/100/100", mockProcessorRouter, 500, []byte("Something went wrong\n"), "text/plain; charset=utf-8"},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", test.URL, nil)
		test.Router.ServeHTTP(w, req)
		if w.Code != test.ExpectedStatus {
			t.Errorf("%s: wrong response code, %#v", test.Name, w.Code)
			continue
		}

		if w.Header()["Content-Type"][0] != test.ExpectedContentType {
			t.Errorf("%s: wrong content type, %#v", test.Name, w.Header()["Content-Type"][0])
		}

		if !reflect.DeepEqual(w.Body.Bytes(), test.ExpectedResponse) {
			t.Errorf("%s: wrong response %#v", test.Name, w.Body.String())
		}
	}

	imageTests := []struct {
		Name             string
		URL              string
		ExpectedStatus   int
		ExpectedResponse []byte
	}{
		// Images
		{"/id/:id/:size", "/id/1/200", 200, readFixture("size")},
		{"/id/:id/:width/:height", "/id/1/200/120", 200, readFixture("width_height")},
		{"/id/:id/:size?blur", "/id/1/200?blur", 200, readFixture("blur")},
		{"/id/:id/:width/:height?blur", "/id/1/200/200?blur", 200, readFixture("blur")},
		{"/id/:id/:size?grayscale", "/id/1/200?grayscale", 200, readFixture("grayscale")},
		{"/id/:id/:width/:height?grayscale", "/id/1/200/200?grayscale", 200, readFixture("grayscale")},
		{"/id/:id/:size?blur&grayscale", "/id/1/200?blur&grayscale", 200, readFixture("all")},
		{"/id/:id/:width/:height?blur&grayscale", "/id/1/200/200?blur&grayscale", 200, readFixture("all")},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400", 200, readFixture("max_allowed")},
	}

	for _, test := range imageTests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", test.URL, nil)
		router.ServeHTTP(w, req)
		if w.Code != test.ExpectedStatus {
			t.Errorf("%s: wrong response code, %#v", test.Name, w.Code)
			continue
		}

		if w.Header()["Content-Type"][0] != "image/jpeg" {
			t.Errorf("%s: wrong content type, %#v", test.Name, w.Header()["Content-Type"][0])
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
		{"/:size", "/200", "/id/1/200/200"},
		{"/:width/:height", "/200/300", "/id/1/200/300"},
		{"/:size?grayscale", "/200?grayscale", "/id/1/200/200?grayscale"},
		{"/:width/:height?grayscale", "/200/300?grayscale", "/id/1/200/300?grayscale"},
		// Default blur amount
		{"/:size?blur", "/200?blur", "/id/1/200/200?blur=5"},
		{"/:width/:height?blur", "/200/300?blur", "/id/1/200/300?blur=5"},
		{"/:size?grayscale&blur", "/200?grayscale&blur", "/id/1/200/200?grayscale&blur=5"},
		{"/:width/:height?grayscale&blur", "/200/300?grayscale&blur", "/id/1/200/300?grayscale&blur=5"},
		// Custom blur amount
		{"/:size?blur=10", "/200?blur=10", "/id/1/200/200?blur=10"},
		{"/:width/:height?blur=10", "/200/300?blur=10", "/id/1/200/300?blur=10"},
		{"/:size?grayscale&blur=10", "/200?grayscale&blur=10", "/id/1/200/200?grayscale&blur=10"},
		{"/:width/:height?grayscale&blur=10", "/200/300?grayscale&blur=10", "/id/1/200/300?grayscale&blur=10"},
		// Deprecated routes
		{"/g/:size", "/g/200", "/200/200?grayscale"},
		{"/g/:width/:height", "/g/200/300", "/200/300?grayscale"},
		{"/g/:size?blur", "/g/200?blur", "/200/200?grayscale&blur=5"},
		{"/g/:width/:height?blur", "/g/200/300?blur", "/200/300?grayscale&blur=5"},
		{"/g/:size?image=:id", "/g/200?image=1", "/id/1/200/200?grayscale"},
		{"/g/:width/:height?image=:id", "/g/200/300?image=1", "/id/1/200/300?grayscale"},
		// Deprecated query params
		{"/:size?image=:id", "/200?image=1", "/id/1/200/200"},
		{"/:width/:height?image=:id", "/200/300?image=1", "/id/1/200/300"},
		{"/:size?image=:id&grayscale", "/200?image=1&grayscale", "/id/1/200/200?grayscale"},
		{"/:width/:height?image=:id&grayscale", "/200/300?image=1&grayscale", "/id/1/200/300?grayscale"},
		{"/:size?image=:id&blur", "/200?image=1&blur", "/id/1/200/200?blur=5"},
		{"/:width/:height?image=:id&blur", "/200/300?image=1&blur", "/id/1/200/300?blur=5"},
		{"/:size?image=:id&grayscale&blur", "/200?image=1&grayscale&blur", "/id/1/200/200?grayscale&blur=5"},
		{"/:width/:height?image=:id&grayscale&blur", "/200/300?image=1&grayscale&blur", "/id/1/200/300?grayscale&blur=5"},
		// Trailing slashes
		{"/:size", "/200/", "/200"},
		{"/:width/:height", "/200/300/", "/200/300"},
		{"/id/:id/:size/", "/id/1/200/", "/id/1/200"},
		{"/id/:id/:width/:height/", "/id/1/200/120/", "/id/1/200/120"},
	}

	for _, test := range redirectTests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", test.URL, nil)
		router.ServeHTTP(w, req)
		if w.Code != 302 && w.Code != 301 {
			t.Errorf("%s: wrong response code, %#v", test.Name, w.Code)
			continue
		}

		if w.Header()["Location"][0] != test.ExpectedURL {
			t.Errorf("%s: wrong redirect %s", test.Name, w.Header()["Location"][0])
		}
	}
}
