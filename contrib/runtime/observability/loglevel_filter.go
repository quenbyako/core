package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

var _ sdklog.Processor = (*levelFilter)(nil)

// Exporter writes JSON-encoded log records to an [io.Writer] ([os.Stdout] by default).
// Exporter must be created with [New].
type levelFilter struct {
	level log.Severity

	inner sdklog.Processor
}

// New creates an [Exporter].
func limitLevel(level slog.Level, inner sdklog.Processor) *levelFilter {
	const sevOffset = slog.Level(log.SeverityDebug) - slog.LevelDebug

	return &levelFilter{
		level: log.Severity(level + sevOffset),
		inner: inner,
	}
}

// ForceFlush implements [log.Processor].
func (e *levelFilter) ForceFlush(ctx context.Context) error {
	return e.inner.ForceFlush(ctx)
}

// OnEmit implements [log.Processor].
func (e *levelFilter) OnEmit(ctx context.Context, record *sdklog.Record) error {
	return e.inner.OnEmit(ctx, record)
}

// Shutdown implements [log.Processor].
func (e *levelFilter) Shutdown(ctx context.Context) error {
	return e.inner.Shutdown(ctx)
}

// Export exports log records to writer.
func (e *levelFilter) Enabled(ctx context.Context, param sdklog.EnabledParameters) bool {
	return param.Severity >= e.level && e.inner.Enabled(ctx, param)
}
