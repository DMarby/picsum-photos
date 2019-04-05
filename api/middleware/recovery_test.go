package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DMarby/picsum-photos/api/middleware"
	"github.com/DMarby/picsum-photos/logger"
	"go.uber.org/zap"
)

func TestRecovery(t *testing.T) {
	log := logger.New(zap.FatalLevel)
	defer log.Sync()

	ts := httptest.NewServer(middleware.Recovery(log, http.HandlerFunc(panicHandler)))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code %#v", res.StatusCode)
	}
}

func panicHandler(rw http.ResponseWriter, req *http.Request) {
	panic("panicking handler")
}
