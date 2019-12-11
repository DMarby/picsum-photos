package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DMarby/picsum-photos/internal/api/handler"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		Name            string
		Method          string
		ExpectedStatus  int
		Headers         map[string]string
		ExpectedHeaders map[string]string
	}{
		{
			Name:           "sets correct headers for non-option requests",
			Method:         "GET",
			ExpectedStatus: http.StatusOK,
			Headers: map[string]string{
				"Origin": "http://www.example.com/",
			},
			ExpectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":   "*",
				"Access-Control-Expose-Headers": "Link, Picsum-ID",
			},
		},
		{
			Name:           "bad request with missing request method header",
			Method:         "OPTIONS",
			ExpectedStatus: http.StatusBadRequest,
			Headers: map[string]string{
				"Origin": "http://www.example.com/",
			},
		},
		{
			Name:           "bad request with wrong request method header",
			Method:         "OPTIONS",
			ExpectedStatus: http.StatusMethodNotAllowed,
			Headers: map[string]string{
				"Origin":                        "http://www.example.com/",
				"Access-Control-Request-Method": "POST",
			},
		},
		{
			Name:           "responds correctly to option request",
			Method:         "OPTIONS",
			ExpectedStatus: http.StatusOK,
			Headers: map[string]string{
				"Origin":                         "http://www.example.com/",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "foobar",
			},
			ExpectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "foobar",
			},
		},
	}

	for _, test := range tests {
		r, err := http.NewRequest(test.Method, "http://www.example.com/", nil)
		if err != nil {
			t.Errorf("%s: %s", test.Name, err)
			continue
		}

		for header, value := range test.Headers {
			r.Header.Set(header, value)
		}

		rr := httptest.NewRecorder()
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		handler.CORS(testHandler).ServeHTTP(rr, r)

		if rr.Code != test.ExpectedStatus {
			t.Errorf("%s: wrong response code, %#v", test.Name, rr.Code)
			continue
		}

		if test.ExpectedHeaders != nil {
			for header, _ := range rr.HeaderMap {
				if _, ok := test.ExpectedHeaders[header]; !ok {
					t.Errorf("%s: unknown header found in response: %s", test.Name, header)
					break
				}
			}

			for expectedHeader, expectedValue := range test.ExpectedHeaders {
				headerValue := rr.Header().Get(expectedHeader)
				if headerValue != expectedValue {
					t.Errorf("%s: wrong header value for %s, %#v", test.Name, expectedHeader, headerValue)
				}
			}
		}
	}
}
