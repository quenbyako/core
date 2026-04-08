package core

import (
	"context"
	"os"
	"os/signal"
)

// BuildContext constructs a root application context annotated with identity
// ([AppName]), version ([AppVersion]) and pipeline I/O ([Pipeline]), and
// automatically wired to OS interrupt signals. The returned cancel function
// MUST be invoked by the caller to release signal resources.
//
// Cancellation Sources:
//   - Incoming SIGINT / SIGKILL ([os.Interrupt], [os.Kill]) trigger context
//     cancellation for graceful shutdown.
//   - Manual invocation of the returned cancel function.
//
// The supplied [Pipeline] is stored for later retrieval via [PipelinesFromContext].
// Prefer passing explicit version / name values; fallback defaults remain
// available through helper extraction funcs.
func BuildContext(
	name AppName,
	version AppVersion,
	pipeline Pipeline,
) (
	ctx context.Context,
	cancel context.CancelFunc,
) {
	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	ctx = WithAppName(ctx, name)
	ctx = WithVersion(ctx, version)
	ctx = WithPipelines(ctx, pipeline)

	return ctx, cancel
}

type ctxAppNameKey struct{}

// WithAppName returns a derived context carrying the provided [AppName].
// It is a lightweight convenience used during application startup to
// annotate the root context with identity metadata that downstream code
// can retrieve via [AppNameFromContext].
//
// The stored value is immutable.
func WithAppName(ctx context.Context, v AppName) context.Context {
	return context.WithValue(ctx, ctxAppNameKey{}, v)
}

// AppNameFromContext extracts an [AppName] previously attached with
// [WithAppName]. When no value is present a stable default is returned and
// the boolean is false, allowing callers to distinguish between implicit
// and explicit identity.
func AppNameFromContext(ctx context.Context) (AppName, bool) {
	if v, ok := ctx.Value(ctxAppNameKey{}).(AppName); ok {
		return v, true
	}

	return defaultAppName(), false
}

type ctxPipelineKey struct{}

// WithPipelines stores a [Pipeline] in a derived context for later retrieval.
// Use [PipelinesFromContext] to extract it; if absent, a cached default is
// provided.
func WithPipelines(ctx context.Context, p Pipeline) context.Context {
	return context.WithValue(ctx, ctxPipelineKey{}, p)
}

// PipelinesFromContext retrieves a [Pipeline] previously attached with
// [WithPipelines]. The boolean reports whether an explicit value was set.
// When false a lazily-created default wrapping the process stdio streams
// is returned.
func PipelinesFromContext(ctx context.Context) (Pipeline, bool) {
	if p, ok := ctx.Value(ctxPipelineKey{}).(Pipeline); ok {
		return p, true
	}

	return defaultPipeline(), false
}

type ctxVersionKey struct{}

// WithVersion attaches an [AppVersion] to a derived context for later
// retrieval via [VersionFromContext]. The stored value is immutable.
func WithVersion(ctx context.Context, v AppVersion) context.Context {
	return context.WithValue(ctx, ctxVersionKey{}, v)
}

// VersionFromContext extracts an [AppVersion] previously attached with
// [WithVersion]. When absent it returns a lazily constructed default and
// false. Callers can use the boolean to differentiate explicit vs.
// fallback version data.
func VersionFromContext(ctx context.Context) (AppVersion, bool) {
	if v, ok := ctx.Value(ctxVersionKey{}).(AppVersion); ok {
		return v, true
	}

	return defaultVersion(), false
}
