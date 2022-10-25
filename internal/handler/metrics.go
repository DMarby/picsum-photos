package handler

import (
	"expvar"
	"net/http"
	"strconv"

	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"tailscale.com/tsweb"
)

var (
	httpRequestsInFlight = expvar.NewInt("gauge_http_requests_in_flight")

	registry                   = prometheus.NewRegistry()
	httpRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Buckets:   durationBuckets,
	}, []string{"path", "code"})
)

var durationBuckets = []float64{
	.01,  // 10 ms
	.025, // 25 ms
	.05,  // 50 ms
	.1,   // 100 ms
	.25,  // 250 ms
	.5,   // 500 ms
	1,    // 1 s
	2.5,  // 2.5 s
	3,    // 3 s
	4,    // 4 s
	5,    // 5 s
	10,   // 10 s
	30,   // 30 s
	45,   // 45 s
}

func init() {
	registry.MustRegister(httpRequestDurationSeconds)
}

func VarzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	expvar.Do(func(kv expvar.KeyValue) {
		tsweb.WritePrometheusExpvar(w, kv)
	})

	mfs, _ := registry.Gather()
	enc := expfmt.NewEncoder(w, expfmt.FmtText)

	for _, mf := range mfs {
		enc.Encode(mf)
	}

	if closer, ok := enc.(expfmt.Closer); ok {
		closer.Close()
	}
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

		// Exclude metrics for certain statuscodes to reduce cardinality
		switch respMetrics.Code {
		// Only set by mux's strict slash redirect
		case http.StatusMovedPermanently:
			return
		// Produced by http.ServeFile when serving the static assets for the website
		case http.StatusPartialContent, http.StatusNotModified, http.StatusRequestedRangeNotSatisfiable:
			return
		}

		histogram := httpRequestDurationSeconds.WithLabelValues(route, strconv.Itoa(respMetrics.Code))
		histogram.Observe(respMetrics.Duration.Seconds())
	})
}
