package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

//nolint:grouper // implements one interface only
var _ sdklog.Processor = (*levelFilter)(nil)

type levelFilter struct {
	inner sdklog.Processor
	level log.Severity
}

// limitLevel creates a log level filter.
func limitLevel(level slog.Level, inner sdklog.Processor) *levelFilter {
	const sevOffset = slog.Level(log.SeverityDebug) - slog.LevelDebug

	return &levelFilter{
		level: log.Severity(level + sevOffset),
		inner: inner,
	}
}

// ForceFlush implements [log.Processor].
func (e *levelFilter) ForceFlush(ctx context.Context) error {
	//nolint:wrapcheck // levelFilter must be silent
	return e.inner.ForceFlush(ctx)
}

// OnEmit implements [log.Processor].
func (e *levelFilter) OnEmit(ctx context.Context, record *sdklog.Record) error {
	//nolint:wrapcheck // levelFilter must be silent
	return e.inner.OnEmit(ctx, record)
}

// Shutdown implements [log.Processor].
func (e *levelFilter) Shutdown(ctx context.Context) error {
	//nolint:wrapcheck // levelFilter must be silent
	return e.inner.Shutdown(ctx)
}

// Export exports log records to writer.
func (e *levelFilter) Enabled(ctx context.Context, param sdklog.EnabledParameters) bool {
	return param.Severity >= e.level && e.inner.Enabled(ctx, param)
}
