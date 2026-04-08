package core

import (
	"go.opentelemetry.io/otel/log"
	noopLog "go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric"
	noopMetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	noopTrace "go.opentelemetry.io/otel/trace/noop"
)

// Metrics bundles logging, tracing and metrics emission capabilities into a
// single optional interface exposed by ObservabilityAppContext. It embeds
// slog.Handler for structured logging plus OTel tracer and meter providers.
// Implementations SHOULD be safe for concurrent use by multiple goroutines.
type Metrics interface {
	log.LoggerProvider
	trace.TracerProvider
	metric.MeterProvider
}

// NoopMetrics returns a Metrics implementation that discards all log records
// and uses no-op tracer / meter providers. This is a lightweight default for
// tests or commands that do not yet wire observability features.
//
//nolint:ireturn // returns interface on intention.
func NoopMetrics() Metrics {
	return &noopMetrics{
		TracerProvider: noopTrace.NewTracerProvider(),
		MeterProvider:  noopMetric.NewMeterProvider(),
		LoggerProvider: noopLog.NewLoggerProvider(),
	}
}

type noopMetrics struct {
	trace.TracerProvider
	metric.MeterProvider
	log.LoggerProvider
}
