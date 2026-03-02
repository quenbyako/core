package observability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/quenbyako/core"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	noopMetric "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	noopTrace "go.opentelemetry.io/otel/trace/noop"
)

type metrics struct {
	slog.Handler
	trace.TracerProvider
	metric.MeterProvider
}

type newParams struct {
	logWriter    io.Writer
	otlpAddr     *url.URL
	otlpMetadata map[string]string
	metricReader sdkmetric.Reader
	hostname     string
	appVersion   core.AppVersion
	logLevel     slog.Level
}

func (p *newParams) validate() error {
	if p.logWriter == nil {
		return errors.New("log writer is nil")
	}

	return nil
}

type NewOption func(*newParams)

func WithLogWriter(writer io.Writer) NewOption {
	return func(m *newParams) { m.logWriter = writer }
}

func WithLogLevel(level slog.Level) NewOption {
	return func(m *newParams) { m.logLevel = level }
}

func WithOtelAddr(otelAddr *url.URL) NewOption {
	return func(m *newParams) { m.otlpAddr = otelAddr }
}

func WithOtlpMetadata(metadata map[string]string) NewOption {
	return func(m *newParams) { m.otlpMetadata = metadata }
}

func WithHostname(hostname string) NewOption {
	return func(m *newParams) { m.hostname = hostname }
}

func WithMetricReader(reader sdkmetric.Reader) NewOption {
	return func(m *newParams) { m.metricReader = reader }
}

// New creates a new observability Metrics instance
//
//nolint:ireturn // returns interface on intention.
func New(ctx context.Context, opts ...NewOption) (core.Metrics, error) {
	appName, _ := core.AppNameFromContext(ctx)
	version, _ := core.VersionFromContext(ctx)

	params := newParams{
		appVersion: version,
		logWriter:  io.Discard,
		logLevel:   slog.LevelInfo,
		otlpAddr:   nil,
		hostname:   "",
	}
	for _, opt := range opts {
		opt(&params)
	}

	if err := params.validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// service.name             unknown_service:cynosure
	// telemetry.sdk.language   go
	// telemetry.sdk.name       opentelemetry
	// telemetry.sdk.version    1.4.0
	// schemaURL https://opentelemetry.io/schemas/1.39.0

	const (
		serviceNameKey = attribute.Key("service.name")
		serviceVersionKey = attribute.Key("service.version")
	)

	appResource, err := resource.Merge(
		resource.Default(),
		// using schemaless to omit semconv differences.
		// TODO: need to make custom resource creator for otel.
		resource.NewSchemaless(
			serviceNameKey.String(ignoreError(appName.Name())),
			serviceVersionKey.String(ignoreError(version.VersionCommit())),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTel resource: %w", err)
	}

	constantAttrs := []slog.Attr{
		slog.String("service_name", ignoreError(appName.Name())+"@"+version.String()),
		slog.String("hostname", params.hostname),
	}

	logHandler := slog.NewJSONHandler(params.logWriter, &slog.HandlerOptions{
		Level: params.logLevel,
		// anything that is lower info, but not included
		AddSource:   params.logLevel < slog.LevelInfo-1,
		ReplaceAttr: nil,
	}).WithAttrs(constantAttrs)

	tracerProvider, err := newTraceProvider(ctx, params.otlpAddr, params.otlpMetadata, appResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace provider: %w", err)
	}

	var meterProvider metric.MeterProvider = noopMetric.NewMeterProvider()
	if params.metricReader != nil {
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(appResource),
			sdkmetric.WithReader(params.metricReader),
		)
	}

	return &metrics{
		Handler:        logHandler,
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}, nil
}

// newTraceProvider creates a new trace.TracerProvider based on the provided address.
//
//nolint:ireturn // returns interface on intention.
func newTraceProvider(
	ctx context.Context,
	addr *url.URL,
	metadata map[string]string,
	appResource *resource.Resource,
) (
	trace.TracerProvider,
	error,
) {
	if addr == nil {
		return noopTrace.NewTracerProvider(), nil
	}

	var (
		exporter sdktrace.SpanExporter
		err      error
	)

	switch scheme := addr.Scheme; scheme {
	case "http", "https":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpointURL(addr.String()),
		}

		if scheme == "https" {
			opts = append(opts, otlptracehttp.WithTLSClientConfig(nil))
		}
		if len(metadata) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(metadata))
		}

		exporter, err = otlptracehttp.New(ctx, opts...)

	case "grpc":
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(addr.Host),
			otlptracegrpc.WithInsecure(),
		}

		if len(metadata) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(metadata))
		}

		exporter, err = otlptracegrpc.New(ctx, opts...)

	default:
		return nil, fmt.Errorf("unsupported trace exporter protocol: %s", scheme)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
		),
		sdktrace.WithResource(appResource),
	), nil
}

func ignoreError[T any, E any](v T, _ E) T { return v }
