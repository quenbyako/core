package core

import (
	"context"
	"log/slog"

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
	slog.Handler
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
	}
}

var _ slog.Handler = (*noopMetrics)(nil) //nolint:grouper // type check

type noopMetrics struct {
	trace.TracerProvider
	metric.MeterProvider
}

func (n *noopMetrics) Enabled(context.Context, slog.Level) bool  { return false }
func (n *noopMetrics) Handle(context.Context, slog.Record) error { return nil }
func (n *noopMetrics) WithAttrs(attrs []slog.Attr) slog.Handler  { return n }
func (n *noopMetrics) WithGroup(name string) slog.Handler        { return n }
