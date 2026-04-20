// Package observability provides OTel-based metrics, tracing, and logging for the core framework.
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
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	noopTrace "go.opentelemetry.io/otel/trace/noop"
)

const (
	protocolHTTP  = "http"
	protocolHTTPS = "https"
	protocolGRPC  = "grpc"
)

type metrics struct {
	log.LoggerProvider
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

func defaultParams(ctx context.Context, opts ...NewOption) newParams {
	appName, _ := core.AppNameFromContext(ctx)
	version, _ := core.VersionFromContext(ctx)

	params := newParams{
		logWriter:    io.Discard,
		otlpAddr:     nil,
		otlpMetadata: nil,
		metricReader: nil,
		hostname:     "",
		appVersion:   version,
		logLevel:     slog.LevelInfo,
	}

	for _, opt := range opts {
		opt(&params)
	}

	_ = appName

	return params
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	appName, _ := core.AppNameFromContext(ctx)
	version, _ := core.VersionFromContext(ctx)

	const (
		serviceNameKey    = attribute.Key("service.name")
		serviceVersionKey = attribute.Key("service.version")
	)

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			serviceNameKey.String(ignoreError(appName.Name())),
			serviceVersionKey.String(ignoreError(version.VersionCommit())),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("observability: failed to merge resources: %w", err)
	}

	return res, nil
}

// New creates a new observability [core.Metrics] instance.
//
//nolint:ireturn // returns interface on intention.
func New(ctx context.Context, opts ...NewOption) (core.Metrics, error) {
	params := defaultParams(ctx, opts...)
	if err := params.validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	appResource, err := newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTel resource: %w", err)
	}

	logProvider, err := newLogProvider(ctx, &params, appResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create log provider: %w", err)
	}

	tracerProvider, err := newTraceProvider(ctx, &params, appResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace provider: %w", err)
	}

	meterProvider, err := newMeterProvider(ctx, &params, appResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter provider: %w", err)
	}

	return &metrics{
		LoggerProvider: logProvider,
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}, nil
}

//nolint:ireturn // returns interface on intention.
func newTraceExporter(
	ctx context.Context,
	addr *url.URL,
	metadata map[string]string,
) (sdktrace.SpanExporter, error) {
	switch scheme := addr.Scheme; scheme {
	case protocolHTTP, protocolHTTPS:
		return newHTTPTraceExporter(ctx, addr, metadata)
	case protocolGRPC:
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(addr.Host),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithHeaders(metadata),
		)
		if err != nil {
			return nil, fmt.Errorf("observability: failed to create gRPC trace exporter: %w", err)
		}

		return exporter, nil
	default:
		return nil, fmt.Errorf("unsupported trace exporter protocol: %s", scheme)
	}
}

//nolint:ireturn // returns interface on intention.
func newHTTPTraceExporter(
	ctx context.Context,
	addr *url.URL,
	metadata map[string]string,
) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(addr.Host),
	}

	switch addr.Scheme {
	case protocolHTTP:
		opts = append(opts, otlptracehttp.WithInsecure())
	case protocolHTTPS:
		opts = append(opts, otlptracehttp.WithTLSClientConfig(nil))
	}

	if addr.Path != "" && addr.Path != "/" {
		opts = append(opts, otlptracehttp.WithURLPath(addr.Path))
	}

	if len(metadata) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(metadata))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("observability: failed to create HTTP trace exporter: %w", err)
	}

	return exporter, nil
}

//nolint:ireturn // returns interface on intention.
func newTraceProvider(
	ctx context.Context,
	params *newParams,
	appResource *resource.Resource,
) (trace.TracerProvider, error) {
	if params.otlpAddr == nil {
		return noopTrace.NewTracerProvider(), nil
	}

	exporter, err := newTraceExporter(ctx, params.otlpAddr, params.otlpMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
		),
		sdktrace.WithResource(appResource),
	), nil
}

//nolint:ireturn // returns interface on intention.
func newLogExporter(
	ctx context.Context,
	addr *url.URL,
) (sdklog.Exporter, error) {
	switch scheme := addr.Scheme; scheme {
	case protocolHTTP, protocolHTTPS:
		return newHTTPLogExporter(ctx, addr)
	case protocolGRPC:
		exporter, err := otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(addr.Host),
			otlploggrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("observability: failed to create gRPC log exporter: %w", err)
		}

		return exporter, nil
	default:
		return nil, fmt.Errorf("unsupported log exporter protocol: %s", scheme)
	}
}

//nolint:ireturn // returns interface on intention.
func newHTTPLogExporter(
	ctx context.Context,
	addr *url.URL,
) (sdklog.Exporter, error) {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(addr.Host),
	}

	switch addr.Scheme {
	case protocolHTTP:
		opts = append(opts, otlploghttp.WithInsecure())
	case protocolHTTPS:
		opts = append(opts, otlploghttp.WithTLSClientConfig(nil))
	}

	if addr.Path != "" && addr.Path != "/" {
		opts = append(opts, otlploghttp.WithURLPath(addr.Path))
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("observability: failed to create HTTP log exporter: %w", err)
	}

	return exporter, nil
}

//nolint:ireturn // returns interface on intention.
func newLogProvider(
	ctx context.Context,
	params *newParams,
	appResource *resource.Resource,
) (log.LoggerProvider, error) {
	opts := []sdklog.LoggerProviderOption{
		sdklog.WithResource(appResource),
	}

	if params.otlpAddr != nil {
		exporter, err := newLogExporter(ctx, params.otlpAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create log exporter: %w", err)
		}

		proc := sdklog.NewBatchProcessor(exporter)
		opts = append(opts, sdklog.WithProcessor(limitLevel(params.logLevel, proc)))
	}

	stderrLogger, err := stdoutlog.New(stdoutlog.WithWriter(params.logWriter))
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr logger: %w", err)
	}

	stderrProc := sdklog.NewSimpleProcessor(stderrLogger)
	opts = append(opts, sdklog.WithProcessor(limitLevel(params.logLevel, stderrProc)))

	return sdklog.NewLoggerProvider(opts...), nil
}

//nolint:ireturn // returns interface on intention.
func newMeterExporter(
	ctx context.Context,
	addr *url.URL,
) (sdkmetric.Exporter, error) {
	switch scheme := addr.Scheme; scheme {
	case protocolHTTP, protocolHTTPS:
		return newHTTPMeterExporter(ctx, addr)
	case protocolGRPC:
		exporter, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(addr.Host),
			otlpmetricgrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("observability: failed to create gRPC metric exporter: %w", err)
		}

		return exporter, nil
	default:
		return nil, fmt.Errorf("unsupported metric exporter protocol: %s", scheme)
	}
}

//nolint:ireturn // returns interface on intention.
func newHTTPMeterExporter(
	ctx context.Context,
	addr *url.URL,
) (sdkmetric.Exporter, error) {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(addr.Host),
	}

	switch addr.Scheme {
	case protocolHTTP:
		opts = append(opts, otlpmetrichttp.WithInsecure())
	case protocolHTTPS:
		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(nil))
	}

	if addr.Path != "" && addr.Path != "/" {
		opts = append(opts, otlpmetrichttp.WithURLPath(addr.Path))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("observability: failed to create HTTP metric exporter: %w", err)
	}

	return exporter, nil
}

//nolint:ireturn // returns interface on intention.
func newMeterProvider(
	ctx context.Context,
	params *newParams,
	appResource *resource.Resource,
) (metric.MeterProvider, error) {
	opts := []sdkmetric.Option{
		sdkmetric.WithResource(appResource),
	}

	if params.otlpAddr != nil {
		exporter, err := newMeterExporter(ctx, params.otlpAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}

		opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
	}

	if params.metricReader != nil {
		opts = append(opts, sdkmetric.WithReader(params.metricReader))
	}

	return sdkmetric.NewMeterProvider(opts...), nil
}

func ignoreError[T any, E any](v T, _ E) T { return v }
