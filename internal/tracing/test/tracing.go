package test

import (
	"context"

	"github.com/DMarby/picsum-photos/internal/logger"
	"github.com/DMarby/picsum-photos/internal/tracing"
	"go.opentelemetry.io/otel/trace"
)

func Tracer(log *logger.Logger) *tracing.Tracer {
	tp := trace.NewNoopTracerProvider()
	return &tracing.Tracer{
		ServiceName:    "test",
		Log:            log,
		TracerProvider: tp,
		ShutdownFunc: func(context.Context) error {
			return nil
		},
		TracerInstance: tp.Tracer("test"),
	}
}
