package observability

import (
	"context"
	"log/slog"
	"slices"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

var _ sdklog.Exporter = (*levelFilter)(nil)

// Exporter writes JSON-encoded log records to an [io.Writer] ([os.Stdout] by default).
// Exporter must be created with [New].
type levelFilter struct {
	level log.Severity

	inner sdklog.Exporter
}

// New creates an [Exporter].
func limitLevel(level slog.Level, inner sdklog.Exporter) *levelFilter {
	const sevOffset = slog.Level(log.SeverityDebug) - slog.LevelDebug

	return &levelFilter{
		level: log.Severity(level + sevOffset),
		inner: inner,
	}
}

// Export exports log records to writer.
func (e *levelFilter) Export(ctx context.Context, records []sdklog.Record) error {
	slices.DeleteFunc(records, func(r sdklog.Record) bool {
		return r.Severity() < e.level
	})

	return e.inner.Export(ctx, records)
}

// Shutdown shuts down the Exporter.
// Calls to Export will perform no operation after this is called.
func (e *levelFilter) Shutdown(ctx context.Context) error {
	return e.inner.Shutdown(ctx)
}

// ForceFlush performs no action.
func (e *levelFilter) ForceFlush(ctx context.Context) error {
	return e.inner.ForceFlush(ctx)
}
