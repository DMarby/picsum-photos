package tracing

import (
	"context"
	"fmt"
	"log"

	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/go-logr/stdr"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerIdentifier = "github.com/DMarby/picsum-photos/internal/tracing"

type Tracer struct {
	ServiceName string
	Log         *logger.Logger

	trace.TracerProvider

	ShutdownFunc   func(context.Context) error
	TracerInstance trace.Tracer
}

func New(ctx context.Context, log *logger.Logger, serviceName string) (*Tracer, error) {
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create opentelemetry grpc exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))),
	)

	// Override the global otel logging
	otel.SetLogger(stdr.New(zap.NewStdLog(log.Desugar())))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Error(err)
	}))

	return &Tracer{
		serviceName,
		log,
		tp,
		tp.Shutdown,
		tp.Tracer(tracerIdentifier),
	}, nil
}

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.TracerInstance.Start(ctx, spanName, opts...)
}

func (t *Tracer) Shutdown(ctx context.Context) {
	if err := t.ShutdownFunc(ctx); err != nil {
		log.Fatal("failed to shutdown tracer: %w", err)
	}
}

func TraceInfo(ctx context.Context) (string, string) {
	traceID := trace.SpanContextFromContext(ctx).TraceID().String()
	spanID := trace.SpanContextFromContext(ctx).SpanID().String()
	return traceID, spanID
}
