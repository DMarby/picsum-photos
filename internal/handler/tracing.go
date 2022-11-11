package handler

import (
	"net/http"

	"github.com/DMarby/picsum-photos/internal/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
)

// Tracer is a handler that adds tracing for handlers
func Tracer(tracer *tracing.Tracer, h http.Handler, routeMatcher RouteMatcher) http.Handler {
	traceHandler := otelhttp.NewHandler(
		h,
		"http",
		otelhttp.WithTracerProvider(tracer),
		otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return routeMatcher.Match(r)
		}),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceHandler.ServeHTTP(w, r)
	})
}
