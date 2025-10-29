package core

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log/slog"

	"github.com/quenbyako/core/internal"
	"github.com/quenbyako/core/secrets"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// RegisterEnvParser registers a parser function for the concrete type T,
// enabling custom environment value decoding for that type.
//
// Usage:
//   - Call only from init() to ensure deterministic, one-time registration.
//   - The registration is global; duplicate registrations for the same T panic.
//   - The lookup key is exactly [reflect.Type]. Pointer forms must be
//     registered separately if they require distinct parsing.
//
// Panics:
//   - If a parser for T is already present.
//
// Concurrency:
//   - Expected to run during startup before other goroutines; no synchronization.
//
// Parser Contract (parseFunc):
//   - Must be pure (no hidden global mutations).
//   - Should not panic on malformed input; return an error instead.
//   - May use context for cancellation or ancillary lookups.
//
// Example:
//
//	func init() {
//	    RegisterEnvParser[MyType](func(ctx context.Context, raw string) (MyType, error) {
//	        var v MyType
//	        if err := v.Unmarshal(raw); err != nil {
//	            return MyType{}, err
//	        }
//	        return v, nil
//	    })
//	}
//
// Hint: extract parser function to separated named function for clarity. It's
// recommended to keep it private.
func RegisterEnvParser[T any](f func(context.Context, string) (T, error)) {
	internal.RegisterEnvParser(f)
}

// EnvParam models a lifecycle-aware configurable entity exposed through
// environment values (e.g., servers, listeners, credentials). The methods are
// invoked in order: [EnvParam.Configure] -> [EnvParam.Acquire] ->
// [EnvParam.Shutdown]. Implementations should be idempotent where feasible and
// release resources in Shutdown.
type EnvParam interface {
	Configure(ctx context.Context, data *ConfigureData) error
	Acquire(ctx context.Context, data *AcquireData) error
	Shutdown(ctx context.Context, data *ShutdownData) error
}

// ConfigureData provides foundational wiring inputs for [EnvParam.Configure].
// Fields may be nil when a capability is absent (e.g., Secrets, Metric).
// Implementations should not mutate shared values.
type ConfigureData struct {
	AppCert tls.Certificate
	Logger  slog.Handler
	Secrets secrets.Engine
	Metric  metric.MeterProvider
	Trace   trace.TracerProvider
	Pool    *x509.CertPool
	Version AppVersion
}

// AcquireData inherits configuration values and allows acquisition logic
// (e.g., binding network listeners). Additional runtime derived fields can
// be layered in future without breaking implementers.
type AcquireData struct {
	ConfigureData
}

// ShutdownData inherits acquisition context for graceful teardown.
// Implementations should attempt best-effort cleanup returning rich
// errors rather than panicking.
type ShutdownData struct {
	AcquireData
}
