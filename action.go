package core

import (
	"context"
	"io"
	"log/slog"
	"net/url"
)

// ActionConfig is the minimal contract every concrete configuration must
// satisfy. The private _ActionConfig method acts as a marker to avoid
// accidental interface satisfaction by unrelated types. Implementations
// typically hold immutable, application-wide tunables (log level, certificate
// paths, metrics endpoint, etc.).
type ActionConfig interface {
	_ActionConfig()

	// log level
	GetLogLevel() slog.Level
	// paths to CA certificates. May return zero slice.
	GetCertPaths() []string
	// path to client certificate
	ClientCertPaths() (cert, key string)
	// secret DSNs
	//
	// TODO: two engines with one protocol? like vault-1:// and vault-2://?
	GetSecretDSNs() map[string]*url.URL
	// OTEL trace endpoint
	GetTraceEndpoint() *url.URL
	// Prometheus metrics address to listen. If nil, metrics export is disabled
	GetMetricsAddr() *url.URL
}

// UnsafeActionConfig is an empty opt-in marker that satisfies [ActionConfig]
// via embedding. Use it when quickly scaffolding a config type; replace with
// explicit methods as requirements grow.
type UnsafeActionConfig struct{}

func (UnsafeActionConfig) _ActionConfig() {}

// UnimplementedActionConfig is a convenient neutral stub returning zero / nil
// values. It is useful for tests, prototypes, or commands that do not yet need
// to surface configuration. All getters return inert defaults (e.g.,
// [slog.LevelInfo], nil slices, nil URLs).
type UnimplementedActionConfig struct {
	UnsafeActionConfig
}

var _ ActionConfig = (*UnimplementedActionConfig)(nil) //nolint:grouper // type check

func (u UnimplementedActionConfig) GetLogLevel() slog.Level             { return slog.LevelInfo }
func (u UnimplementedActionConfig) GetCertPaths() []string              { return nil }
func (u UnimplementedActionConfig) ClientCertPaths() (cert, key string) { return "", "" }
func (u UnimplementedActionConfig) GetSecretDSNs() map[string]*url.URL  { return nil }
func (u UnimplementedActionConfig) GetTraceEndpoint() *url.URL          { return nil }
func (u UnimplementedActionConfig) GetMetricsAddr() *url.URL            { return nil }

// ExitCode represents the process exit status produced by an ActionFunc. The
// uint8 size mirrors conventional POSIX exit ranges (0–255) and communicates
// intent more clearly than a bare int.
type ExitCode uint8

// ActionFunc is the canonical executable signature for an application action or
// subcommand. It receives:
//   - [context.Context]: For cancellation, deadlines, and cross-cutting values.
//   - [AppContext][T]: A strongly typed application context exposing identity and
//     configuration.
type ActionFunc[T ActionConfig] func(ctx context.Context, appCtx AppContext[T]) ExitCode

type AppContext[T ActionConfig] interface {
	// Stable application identifier (not necessarily user-facing).
	Name() AppName
	// Semantic or custom version descriptor
	Version() AppVersion
	// Concrete configuration value of type T.
	//
	// Implementations SHOULD document whether Config() is safe for concurrent
	// use. Immutable configs are preferred to simplify synchronization.
	Config() T
}

func Name[T ActionConfig](ctx AppContext[T]) AppName       { return ctx.Name() }
func Version[T ActionConfig](ctx AppContext[T]) AppVersion { return ctx.Version() }

// Config returns the concrete configuration value (type T) carried by the
// supplied AppContext. It is a thin, inline-able alias for ctx.Config()
// provided for symmetry with the Name and Version helpers and to improve
// readability at call sites.
//
// Characteristics:
//   - Zero overhead: No copying beyond what ctx.Config() itself performs;
//     the generic accessor typically inlines.
//   - Strongly typed: Callers receive the full concrete T, enabling direct
//     field / method access without casts or interface assertions.
//   - Concurrency: Safety of the returned value depends on the AppContext
//     implementation. Prefer immutable configuration structs after
//     initialization.
//
// Usage Example:
//
//	cfg := Config(appCtx)          // obtains T
//	lvl := cfg.GetLogLevel()       // invoke concrete methods directly
//
// Mutability Guidance:
// Modifying cfg is discouraged unless the specific configuration type
// documents that such mutation is safe. Treat configuration as read-only
// in most application code.
//
// Equivalent Call:
//
//	Config(appCtx) == appCtx.Config()
func Config[T ActionConfig](ctx AppContext[T]) T { return ctx.Config() }

type PipelineAppContext[T ActionConfig] interface {
	AppContext[T]

	Stdin() io.Reader
	Stdout() io.Writer
}

func Stdin[T ActionConfig](ctx AppContext[T]) (io.Reader, bool) {
	if v, ok := ctx.(PipelineAppContext[T]); ok {
		return v.Stdin(), ok
	}

	return nil, false
}

func Stdout[T ActionConfig](ctx AppContext[T]) (io.Writer, bool) {
	if v, ok := ctx.(PipelineAppContext[T]); ok {
		return v.Stdout(), ok
	}

	return nil, false
}

type LoggerAppContext[T ActionConfig] interface {
	AppContext[T]

	Log() slog.Handler
}

// Logger attempts to extract a [slog.Handler] logging capability from the provided
// [AppContext]. It performs a single type assertion against [LoggerAppContext].
// Returns (handler, true) when the capability is present, or (nil, false) if the
// context does not supply structured logging.
//
// Semantics:
//   - Absence is not an error; callers should branch on the boolean and degrade
//     gracefully (e.g., use a no-op handler or skip logging).
//   - The returned [slog.Handler] SHOULD be safe for concurrent use; this is an
//     implementation concern of the concrete AppContext.
//
// Example:
//
//	if h, ok := Logger(appCtx); ok {
//	    h.Handle(ctx, slog.Record{ /* ... */ })
//	}
func Logger[T ActionConfig](ctx AppContext[T]) (slog.Handler, bool) {
	if v, ok := ctx.(LoggerAppContext[T]); ok {
		return v.Log(), ok
	}

	return nil, false
}

type ObservabilityAppContext[T ActionConfig] interface {
	AppContext[T]

	Observability() Metrics
}

//nolint:ireturn // returns interface on intention.
func Observability[T ActionConfig](ctx AppContext[T]) (Metrics, bool) {
	if v, ok := ctx.(ObservabilityAppContext[T]); ok {
		return v.Observability(), ok
	}

	return nil, false
}
