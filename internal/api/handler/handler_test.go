package handler_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/DMarby/picsum-photos/internal/api/handler"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		Name                string
		AcceptHeader        string
		ExpectedContentType string
		ExpectedStatus      int
		ExpectedResponse    []byte
		Handler             handler.Handler
	}{
		{"internal server error", "text/html", "text/plain; charset=utf-8", http.StatusInternalServerError, []byte("Something went wrong\n"), errorHandler},
		{"internal server error json", "application/json", "application/json", http.StatusInternalServerError, []byte("{\"error\":\"Something went wrong\"}\n"), errorHandler},
		{"bad request", "text/html", "text/plain; charset=utf-8", http.StatusBadRequest, []byte("Bad request test\n"), badRequestHandler},
		{"bad request json", "application/json", "application/json", http.StatusBadRequest, []byte("{\"error\":\"Bad request test\"}\n"), badRequestHandler},
	}

	for _, test := range tests {
		ts := httptest.NewServer(handler.Handler(test.Handler))
		defer ts.Close()

		req, err := http.NewRequest("GET", ts.URL, nil)
		if err != nil {
			t.Errorf("%s: %s", test.Name, err)
			continue
		}

		req.Header.Set("Accept", test.AcceptHeader)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("%s: %s", test.Name, err)
			continue
		}

		defer res.Body.Close()

		if res.StatusCode != test.ExpectedStatus {
			t.Errorf("%s: wrong response code, %#v", test.Name, res.StatusCode)
			continue
		}

		contentType := res.Header.Get("Content-Type")
		if contentType != test.ExpectedContentType {
			t.Errorf("%s: wrong content type, %#v", test.Name, contentType)
			continue
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("%s: %s", test.Name, err)
			continue
		}

		if !reflect.DeepEqual(body, test.ExpectedResponse) {
			t.Errorf("%s: wrong response %s", test.Name, body)
		}
	}

}

func errorHandler(rw http.ResponseWriter, req *http.Request) *handler.Error {
	return handler.InternalServerError()
}

func badRequestHandler(rw http.ResponseWriter, req *http.Request) *handler.Error {
	return handler.BadRequest("Bad request test")
}
