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
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	noopTrace "go.opentelemetry.io/otel/trace/noop"
)

type metrics struct {
	log.LoggerProvider
	trace.TracerProvider
	metric.MeterProvider
}

type newParams struct {
	logWriter    io.Writer
	otelAddr     *url.URL
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
	return func(m *newParams) { m.otelAddr = otelAddr }
}

func WithHostname(hostname string) NewOption {
	return func(m *newParams) { m.hostname = hostname }
}

func WithMetricReader(reader sdkmetric.Reader) NewOption {
	return func(m *newParams) { m.metricReader = reader }
}

// New creates a new observability [core.Metrics] instance
//
//nolint:ireturn // returns interface on intention.
func New(ctx context.Context, opts ...NewOption) (core.Metrics, error) {
	appName, _ := core.AppNameFromContext(ctx)
	version, _ := core.VersionFromContext(ctx)

	params := newParams{
		appVersion: version,
		logWriter:  io.Discard,
		logLevel:   slog.LevelInfo,
		otelAddr:   nil,
		hostname:   "",
	}
	for _, opt := range opts {
		opt(&params)
	}

	if err := params.validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	appResource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ignoreError(appName.Name())),
			semconv.ServiceVersion(ignoreError(version.VersionCommit())),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTel resource: %w", err)
	}

	logProvider, err := newLogProvider(ctx, params.logWriter, params.logLevel, params.otelAddr, appResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create log provider: %w", err)
	}

	tracerProvider, err := newTraceProvider(ctx, params.otelAddr, appResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace provider: %w", err)
	}

	meterProvider, err := newMeterProvider(ctx, params.otelAddr, appResource, params.metricReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter provider: %w", err)
	}

	return &metrics{
		LoggerProvider: logProvider,
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

		exporter, err = otlptracehttp.New(ctx, opts...)

	case "grpc":
		exporter, err = otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(addr.Host),
			otlptracegrpc.WithInsecure(),
		)

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

// newLogProvider creates a new [log.LoggerProvider] based on the provided address.
//
//nolint:ireturn // returns interface on intention.
func newLogProvider(
	ctx context.Context,
	stderr io.Writer,
	level slog.Level,
	addr *url.URL,
	appResource *resource.Resource,
) (
	log.LoggerProvider,
	error,
) {
	opts := []sdklog.LoggerProviderOption{
		sdklog.WithResource(appResource),
	}

	if addr != nil {
		var (
			exporter sdklog.Exporter
			err      error
		)

		switch scheme := addr.Scheme; scheme {
		case "http", "https":
			opts := []otlploghttp.Option{
				otlploghttp.WithEndpointURL(addr.String()),
			}

			if scheme == "https" {
				opts = append(opts, otlploghttp.WithTLSClientConfig(nil))
			}

			exporter, err = otlploghttp.New(ctx, opts...)

		case "grpc":
			exporter, err = otlploggrpc.New(
				ctx,
				otlploggrpc.WithEndpoint(addr.Host),
				otlploggrpc.WithInsecure(),
			)

		default:
			return nil, fmt.Errorf("unsupported log exporter protocol: %s", scheme)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create log exporter: %w", err)
		}

		opts = append(opts,
			sdklog.WithProcessor(sdklog.NewBatchProcessor(limitLevel(level, exporter))),
		)
	}

	stderrLogger, err := stdoutlog.New(
		stdoutlog.WithWriter(stderr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr logger: %w", err)
	}
	opts = append(opts,
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(limitLevel(level, stderrLogger))),
	)

	return sdklog.NewLoggerProvider(opts...), nil
}

func newMeterProvider(
	ctx context.Context,
	addr *url.URL,
	appResource *resource.Resource,
	reader sdkmetric.Reader,
) (
	metric.MeterProvider,
	error,
) {
	opts := []sdkmetric.Option{
		sdkmetric.WithResource(appResource),
	}

	if addr != nil {
		var (
			exporter sdkmetric.Exporter
			err      error
		)

		switch scheme := addr.Scheme; scheme {
		case "http", "https":
			opts := []otlpmetrichttp.Option{
				otlpmetrichttp.WithEndpointURL(addr.String()),
			}

			if scheme == "https" {
				opts = append(opts, otlpmetrichttp.WithTLSClientConfig(nil))
			}

			exporter, err = otlpmetrichttp.New(ctx, opts...)

		case "grpc":
			exporter, err = otlpmetrichttp.New(
				ctx,
				otlpmetrichttp.WithEndpoint(addr.Host),
				otlpmetrichttp.WithInsecure(),
			)

		default:
			return nil, fmt.Errorf("unsupported metric exporter protocol: %s", scheme)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}

		opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
	}

	if reader != nil {
		opts = append(opts, sdkmetric.WithReader(reader))
	}

	return sdkmetric.NewMeterProvider(opts...), nil

}

func ignoreError[T any, E any](v T, _ E) T { return v }
