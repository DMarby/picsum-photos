package handler

import (
	"expvar"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
)

var httpRequestsInFlight = expvar.NewInt("gauge_http_requests_in_flight")
var httpRequestDurationSeconds = &RequestHistogram{
	buckets: make(map[string]*bucket),
}

var durationBuckets = []string{
	"0.01s",
	"0.025s",
	"0.05s",
	"0.1s",
	"0.25s",
	"0.5s",
	"1s",
	"5s",
	"10s",
}

func init() {
	expvar.Publish("http_request_duration_seconds", httpRequestDurationSeconds)
}

// Metrics is a handler that collects performance metrics
func Metrics(h http.Handler, routeMatcher RouteMatcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := routeMatcher.Match(r)

		httpRequestsInFlight.Add(1)
		defer httpRequestsInFlight.Add(-1)

		respMetrics := httpsnoop.CaptureMetricsFn(w, func(ww http.ResponseWriter) {
			h.ServeHTTP(ww, r)
		})

		httpRequestDurationSeconds.Add(route, respMetrics.Code, respMetrics.Duration)
	})
}

type bucket struct {
	m             expvar.Map
	count         expvar.Int
	totalDuration expvar.Float
}

type RequestHistogram struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

func (r *RequestHistogram) Add(path string, code int, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d;%s", code, path)

	b, exists := r.buckets[key]
	if !exists {
		newBucket := &bucket{}
		r.buckets[key], b = newBucket, newBucket
	}

	b.count.Add(1)
	b.totalDuration.Add(duration.Seconds())

	for _, db := range durationBuckets {
		pdb, _ := time.ParseDuration(db)
		if duration <= pdb {
			b.m.Add(strings.Trim(db, "s"), 1)
		}
	}

	b.m.Add("+Inf", 1)
}

func (r *RequestHistogram) WritePrometheus(w io.Writer, prefix string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Fprintf(w, "# TYPE %s histogram\n", prefix)

	for key, b := range r.buckets {
		if code, path, ok := strings.Cut(key, ";"); ok {
			b.m.Do(func(kv expvar.KeyValue) {
				fmt.Fprintf(w, "%s_bucket{path=%q,code=%q,le=%q} %v\n", prefix, path, code, kv.Key, kv.Value)
			})

			fmt.Fprintf(w, "%s_count{path=%q,code=%q} %v\n", prefix, path, code, b.count.String())
			fmt.Fprintf(w, "%s_sum{path=%q,code=%q} %v\n", prefix, path, code, b.totalDuration.String())
		}
	}
}

func (r *RequestHistogram) String() string {
	return ""
}
