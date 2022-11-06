package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/DMarby/picsum-photos/internal/api"
	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/hmac"
	"github.com/DMarby/picsum-photos/internal/logger"
	"go.uber.org/zap"

	fileDatabase "github.com/DMarby/picsum-photos/internal/database/file"
	mockDatabase "github.com/DMarby/picsum-photos/internal/database/mock"

	"testing"
)

const rootURL = "https://example.com"
const imageServiceURL = "https://i.example.com"

func TestAPI(t *testing.T) {
	log := logger.New(zap.FatalLevel)
	defer log.Sync()

	db, _ := fileDatabase.New("../../test/fixtures/file/metadata.json")
	dbMultiple, _ := fileDatabase.New("../../test/fixtures/file/metadata_multiple.json")

	staticPath := "../../web"

	hmac := &hmac.HMAC{
		Key: []byte("test"),
	}

	router := (&api.API{db, log, rootURL, imageServiceURL, staticPath, time.Minute, hmac}).Router()
	paginationRouter := (&api.API{dbMultiple, log, rootURL, imageServiceURL, staticPath, time.Minute, hmac}).Router()
	mockDatabaseRouter := (&api.API{&mockDatabase.Provider{}, log, rootURL, imageServiceURL, staticPath, time.Minute, hmac}).Router()

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
				{
					Image: database.Image{
						ID:     "1",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/1/300/400", rootURL),
				},
				{
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
				"Content-Type":                  "application/json",
				"Link":                          fmt.Sprintf("<%s/v2/list?page=2&limit=30>; rel=\"next\"", rootURL),
				"Cache-Control":                 "no-cache, no-store, must-revalidate",
				"Access-Control-Expose-Headers": "Link",
			},
		},
		{
			Name:           "/v2/list lists images with limit",
			URL:            "/v2/list?limit=1000",
			Router:         paginationRouter,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{
				{
					Image: database.Image{
						ID:     "1",
						Author: "John Doe",
						URL:    "https://picsum.photos",
						Width:  300,
						Height: 400,
					},
					DownloadURL: fmt.Sprintf("%s/id/1/300/400", rootURL),
				},
				{
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
				{
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
				"Content-Type":                  "application/json",
				"Link":                          fmt.Sprintf("<%s/v2/list?page=2&limit=1>; rel=\"next\"", rootURL),
				"Cache-Control":                 "no-cache, no-store, must-revalidate",
				"Access-Control-Expose-Headers": "Link",
			},
		},
		{
			Name:           "/v2/list pagination page 2",
			URL:            "/v2/list?page=2&limit=1",
			Router:         paginationRouter,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{
				{
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
				"Content-Type":                  "application/json",
				"Link":                          fmt.Sprintf("<%s/v2/list?page=1&limit=1>; rel=\"prev\", <%s/v2/list?page=3&limit=1>; rel=\"next\"", rootURL, rootURL),
				"Cache-Control":                 "no-cache, no-store, must-revalidate",
				"Access-Control-Expose-Headers": "Link",
			},
		},
		{
			Name:             "/v2/list pagination page 3",
			URL:              "/v2/list?page=3&limit=1",
			Router:           paginationRouter,
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: marshalJson([]api.ListImage{}),
			ExpectedHeaders: map[string]string{
				"Content-Type":                  "application/json",
				"Link":                          fmt.Sprintf("<%s/v2/list?page=2&limit=1>; rel=\"prev\"", rootURL),
				"Cache-Control":                 "no-cache, no-store, must-revalidate",
				"Access-Control-Expose-Headers": "Link",
			},
		},
		{
			Name:           "Deprecated /list lists images",
			URL:            "/list",
			Router:         router,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: marshalJson([]api.DeprecatedImage{
				{
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
		{
			Name:           "/seed/{seed}/info returns info about an image",
			URL:            "/seed/1/info",
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
		{"invalid size", "/id/1/1/9223372036854775808", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},   // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/9223372036854775808/1", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},   // Number larger then max int size to fail int parsing
		{"invalid size", "/id/1/5500/1", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},                  // Number larger then maxImageSize to fail int parsing
		{"invalid size", "/seed/1/9223372036854775808/1", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}}, // Number larger then maxImageSize to fail int parsing
		{"invalid size", "/9223372036854775808", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},          // Number larger then maxImageSize to fail int parsing
		{"invalid blur amount", "/id/1/100/100?blur=11", router, http.StatusBadRequest, []byte("Invalid blur amount\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid blur amount", "/id/1/100/100?blur=0", router, http.StatusBadRequest, []byte("Invalid blur amount\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"invalid file extension", "/id/1/100/100.png", router, http.StatusBadRequest, []byte("Invalid file extension\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		// Deprecated handler errors
		{"invalid size", "/g/9223372036854775808", router, http.StatusBadRequest, []byte("Invalid size\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}}, // Number larger then max int size to fail int parsing
		// Database errors
		{"List()", "/list", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"List()", "/v2/list", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"GetRandom()", "/200", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"GetRandom()", "/g/200", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"GetRandomWithSeed()", "/seed/1/200", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"Get() database", "/id/1/100/100", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"Get() database", "/g/100?image=1", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
		{"Get() database info", "/id/1/info", mockDatabaseRouter, http.StatusInternalServerError, []byte("Something went wrong\n"), map[string]string{"Content-Type": "text/plain; charset=utf-8", "Cache-Control": "no-cache, no-store, must-revalidate"}},
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

	redirectTests := []struct {
		Name            string
		URL             string
		ExpectedURL     string
		TestCacheHeader bool
		LocalRedirect   bool
	}{
		// /id/:id/:size to <imageServiceURL>/id/:id/:width/:height
		{"/id/:id/:size", "/id/1/200", "/id/1/200/200.jpg", true, false},
		{"/id/:id/:size.jpg", "/id/1/200.jpg", "/id/1/200/200.jpg", true, false},
		{"/id/:id/:size.webp", "/id/1/200.webp", "/id/1/200/200.webp", true, false},
		{"/id/:id/:size?blur", "/id/1/200?blur", "/id/1/200/200.jpg?blur=5", true, false},
		{"/id/:id/:size?blur", "/id/1/200?blur=10", "/id/1/200/200.jpg?blur=10", true, false},
		{"/id/:id/:size?grayscale", "/id/1/200?grayscale", "/id/1/200/200.jpg?grayscale", true, false},
		{"/id/:id/:size?blur&grayscale", "/id/1/200?blur&grayscale", "/id/1/200/200.jpg?blur=5&grayscale", true, false},

		// General
		{"/:size", "/200", "/id/1/200/200.jpg", true, false},
		{"/:width/:height", "/200/300", "/id/1/200/300.jpg", true, false},
		{"/:size.jpg", "/200.jpg", "/id/1/200/200.jpg", true, false},
		{"/:width/:height.jpg", "/200/300.jpg", "/id/1/200/300.jpg", true, false},
		{"/:size.webp", "/200.webp", "/id/1/200/200.webp", true, false},
		{"/:width/:height.webp", "/200/300.webp", "/id/1/200/300.webp", true, false},
		{"/:size?grayscale", "/200?grayscale", "/id/1/200/200.jpg?grayscale", true, false},
		{"/:width/:height?grayscale", "/200/300?grayscale", "/id/1/200/300.jpg?grayscale", true, false},
		// JPG
		{"/id/:id/:width/:height", "/id/1/200/120", "/id/1/200/120.jpg", true, false},
		{"/id/:id/:width/:height.jpg", "/id/1/200/120.jpg", "/id/1/200/120.jpg", true, false},
		{"/id/:id/:width/:height?blur", "/id/1/200/200?blur", "/id/1/200/200.jpg?blur=5", true, false},
		{"/id/:id/:width/:height.jpg?blur", "/id/1/200/200.jpg?blur", "/id/1/200/200.jpg?blur=5", true, false},
		{"/id/:id/:width/:height?grayscale", "/id/1/200/200?grayscale", "/id/1/200/200.jpg?grayscale", true, false},
		{"/id/:id/:width/:height.jpg?grayscale", "/id/1/200/200.jpg?grayscale", "/id/1/200/200.jpg?grayscale", true, false},
		{"/id/:id/:width/:height?blur&grayscale", "/id/1/200/200?blur&grayscale", "/id/1/200/200.jpg?blur=5&grayscale", true, false},
		{"/id/:id/:width/:height.jpg?blur&grayscale", "/id/1/200/200.jpg?blur&grayscale", "/id/1/200/200.jpg?blur=5&grayscale", true, false},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400", "/id/1/300/400.jpg", true, false},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400.jpg", "/id/1/300/400.jpg", true, false},
		{"width/height of 0 returns original image width", "/id/1/0/0", "/id/1/300/400.jpg", true, false},
		{"width/height of 0 returns original image width", "/id/1/0/0.jpg", "/id/1/300/400.jpg", true, false},
		// WebP
		{"/id/:id/:width/:height.webp", "/id/1/200/120.webp", "/id/1/200/120.webp", true, false},
		{"/id/:id/:width/:height.webp?blur", "/id/1/200/200.webp?blur", "/id/1/200/200.webp?blur=5", true, false},
		{"/id/:id/:width/:height.webp?grayscale", "/id/1/200/200.webp?grayscale", "/id/1/200/200.webp?grayscale", true, false},
		{"/id/:id/:width/:height.webp?blur&grayscale", "/id/1/200/200.webp?blur&grayscale", "/id/1/200/200.webp?blur=5&grayscale", true, false},
		{"width/height larger then max allowed but same size as image", "/id/1/300/400.webp", "/id/1/300/400.webp", true, false},
		{"width/height of 0 returns original image width", "/id/1/0/0.webp", "/id/1/300/400.webp", true, false},

		// Default blur amount
		{"/:size?blur", "/200?blur", "/id/1/200/200.jpg?blur=5", true, false},
		{"/:width/:height?blur", "/200/300?blur", "/id/1/200/300.jpg?blur=5", true, false},
		{"/:size?grayscale&blur", "/200?grayscale&blur", "/id/1/200/200.jpg?blur=5&grayscale", true, false},
		{"/:width/:height?grayscale&blur", "/200/300?grayscale&blur", "/id/1/200/300.jpg?blur=5&grayscale", true, false},

		// Custom blur amount
		{"/:size?blur=10", "/200?blur=10", "/id/1/200/200.jpg?blur=10", true, false},
		{"/:width/:height?blur=10", "/200/300?blur=10", "/id/1/200/300.jpg?blur=10", true, false},
		{"/:size?grayscale&blur=10", "/200?grayscale&blur=10", "/id/1/200/200.jpg?blur=10&grayscale", true, false},
		{"/:width/:height?grayscale&blur=10", "/200/300?grayscale&blur=10", "/id/1/200/300.jpg?blur=10&grayscale", true, false},

		// Deprecated routes
		{"/g/:size", "/g/200", "/id/1/200/200.jpg?grayscale", true, false},
		{"/g/:width/:height", "/g/200/300", "/id/1/200/300.jpg?grayscale", true, false},
		{"/g/:size.jpg", "/g/200.jpg", "/id/1/200/200.jpg?grayscale", true, false},
		{"/g/:width/:height.jpg", "/g/200/300.jpg", "/id/1/200/300.jpg?grayscale", true, false},
		{"/g/:size.webp", "/g/200.webp", "/id/1/200/200.webp?grayscale", true, false},
		{"/g/:width/:height.webp", "/g/200/300.webp", "/id/1/200/300.webp?grayscale", true, false},
		{"/g/:size?blur", "/g/200?blur", "/id/1/200/200.jpg?blur=5&grayscale", true, false},
		{"/g/:width/:height?blur", "/g/200/300?blur", "/id/1/200/300.jpg?blur=5&grayscale", true, false},
		{"/g/:size?image=:id", "/g/200?image=1", "/id/1/200/200.jpg?grayscale", true, false},
		{"/g/:width/:height?image=:id", "/g/200/300?image=1", "/id/1/200/300.jpg?grayscale", true, false},
		{"/g/:size.jpg?image=:id", "/g/200.jpg?image=1", "/id/1/200/200.jpg?grayscale", true, false},
		{"/g/:width/:height.jpg?image=:id", "/g/200/300.jpg?image=1", "/id/1/200/300.jpg?grayscale", true, false},
		{"/g/:size.webp?image=:id", "/g/200.webp?image=1", "/id/1/200/200.webp?grayscale", true, false},
		{"/g/:width/:height.webp?image=:id", "/g/200/300.webp?image=1", "/id/1/200/300.webp?grayscale", true, false},

		// Deprecated query params
		{"/:size?image=:id", "/200?image=1", "/id/1/200/200.jpg", true, false},
		{"/:width/:height?image=:id", "/200/300?image=1", "/id/1/200/300.jpg", true, false},
		{"/:size?image=:id&grayscale", "/200?image=1&grayscale", "/id/1/200/200.jpg?grayscale", true, false},
		{"/:width/:height?image=:id&grayscale", "/200/300?image=1&grayscale", "/id/1/200/300.jpg?grayscale", true, false},
		{"/:size?image=:id&blur", "/200?image=1&blur", "/id/1/200/200.jpg?blur=5", true, false},
		{"/:width/:height?image=:id&blur", "/200/300?image=1&blur", "/id/1/200/300.jpg?blur=5", true, false},
		{"/:size?image=:id&grayscale&blur", "/200?image=1&grayscale&blur", "/id/1/200/200.jpg?blur=5&grayscale", true, false},
		{"/:width/:height?image=:id&grayscale&blur", "/200/300?image=1&grayscale&blur", "/id/1/200/300.jpg?blur=5&grayscale", true, false},

		// By seed
		{"/seed/:seed/:size", "/seed/1/200", "/id/1/200/200.jpg", true, false},
		{"/seed/:seed/:size.jpg", "/seed/1/200.jpg", "/id/1/200/200.jpg", true, false},
		{"/seed/:seed/:size.webp", "/seed/1/200.webp", "/id/1/200/200.webp", true, false},
		{"/seed/:seed/:size?blur", "/seed/1/200?blur", "/id/1/200/200.jpg?blur=5", true, false},
		{"/seed/:seed/:size?blur=10", "/seed/1/200?blur=10", "/id/1/200/200.jpg?blur=10", true, false},
		{"/seed/:seed/:size?grayscale", "/seed/1/200?grayscale", "/id/1/200/200.jpg?grayscale", true, false},
		{"/seed/:seed/:size?blur&grayscale", "/seed/1/200?blur&grayscale", "/id/1/200/200.jpg?blur=5&grayscale", true, false},
		{"/seed/:seed/:size?blur=10&grayscale", "/seed/1/200?blur=10&grayscale", "/id/1/200/200.jpg?blur=10&grayscale", true, false},
		{"/seed/:seed/:width/:height", "/seed/1/200/300", "/id/1/200/300.jpg", true, false},
		{"/seed/:seed/:width/:height.jpg", "/seed/1/200/300.jpg", "/id/1/200/300.jpg", true, false},
		{"/seed/:seed/:width/:height.webp", "/seed/1/200/300.webp", "/id/1/200/300.webp", true, false},
		{"/seed/:seed/:width/:height?blur", "/seed/1/200/300?blur", "/id/1/200/300.jpg?blur=5", true, false},
		{"/seed/:seed/:width/:height?blur=10", "/seed/1/200/300?blur=10", "/id/1/200/300.jpg?blur=10", true, false},
		{"/seed/:seed/:width/:height?grayscale", "/seed/1/200/300?grayscale", "/id/1/200/300.jpg?grayscale", true, false},
		{"/seed/:seed/:width/:height?blur&grayscale", "/seed/1/200/300?blur&grayscale", "/id/1/200/300.jpg?blur=5&grayscale", true, false},
		{"/seed/:seed/:width/:height?blur=10&grayscale", "/seed/1/200/300?blur=10&grayscale", "/id/1/200/300.jpg?blur=10&grayscale", true, false},

		// Trailing slashes
		{"/:size/", "/200/", "/200", false, true},
		{"/:width/:height/", "/200/300/", "/200/300", false, true},
		{"/id/:id/:size/", "/id/1/200/", "/id/1/200", false, true},
		{"/id/:id/:width/:height/", "/id/1/200/120/", "/id/1/200/120", false, true},
		{"/seed/:seed/:size/", "/seed/1/200/", "/seed/1/200", false, true},
		{"/seed/:seed/:width/:height/", "/seed/1/200/120/", "/seed/1/200/120", false, true},
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

		expectedURL := test.ExpectedURL
		if !test.LocalRedirect {
			expectedHMAC, err := hmac.Create(test.ExpectedURL)
			if err != nil {
				t.Errorf("%s: hmac error %s", test.Name, err)
				continue
			}

			if strings.Contains(test.ExpectedURL, "?") {
				expectedURL = imageServiceURL + test.ExpectedURL + "&hmac=" + expectedHMAC
			} else {
				expectedURL = imageServiceURL + test.ExpectedURL + "?hmac=" + expectedHMAC
			}
		}

		if location != expectedURL {
			t.Errorf("%s: wrong redirect %s, expected %s", test.Name, location, expectedURL)
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

func readFile(path string) []byte {
	fixture, _ := os.ReadFile(path)
	return fixture
}
