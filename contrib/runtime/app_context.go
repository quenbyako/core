package runtime

import (
	"crypto/x509"
	"io"

	// "github.com/open-feature/go-sdk/openfeature"
	"github.com/quenbyako/core"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type appCtx[T any] struct {
	appName core.AppName
	version core.AppVersion
	config  T

	stdin  io.Reader
	stdout io.Writer
	log    log.LoggerProvider
	metric metric.MeterProvider
	trace  trace.TracerProvider
	// Features       openfeature.IClient
	caCertificates *x509.CertPool

	IsPipeline bool
}

var _ _allTogether[core.UnimplementedActionConfig] = (*appCtx[core.UnimplementedActionConfig])(nil)

type _allTogether[T core.ActionConfig] interface {
	core.AppContext[T]
	core.ObservabilityAppContext[T]
	core.PipelineAppContext[T]
	core.CircuitBreakerAppContext[T]
}

func (a *appCtx[T]) Name() core.AppName       { return a.appName }
func (a *appCtx[T]) Version() core.AppVersion { return a.version }
func (a *appCtx[T]) Config() T                { return a.config }
func (a *appCtx[T]) Observability() core.Metrics {
	return appObservability{
		MeterProvider:  a.metric,
		TracerProvider: a.trace,
		LoggerProvider: a.log,
	}
}
func (a *appCtx[T]) Stdin() io.Reader  { return a.stdin }
func (a *appCtx[T]) Stdout() io.Writer { return a.stdout }

type appObservability struct {
	metric.MeterProvider
	trace.TracerProvider
	log.LoggerProvider
}

var _ core.Metrics = (*appObservability)(nil)

func (a *appCtx[T]) CircuitBreaker(scope string) core.CircuitBreaker {
	return core.NoopCircuitBreaker()
}
