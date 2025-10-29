// Copyright (c) 2025. All rights reserved.
// PRIVATE: Internal use only. Redistribution, public disclosure, or any
// form of distribution is strictly prohibited without prior written consent.
// Contact: <EMAIL>

// Package core supplies foundational, capability-oriented abstractions for
// building composable CLI and service applications on top of strongly typed,
// generic configuration values. The design emphasizes:
//   - Progressive enhancement: Optional features (I/O, logging, metrics) surface
//     through small extension interfaces rather than a monolithic context.
//   - Type safety via generics: Actions receive their full concrete configuration
//     type (T) without casting, reflection, or map indirection.
//   - Explicit capability discovery: Helper probes return (value, ok) allowing
//     graceful degradation when a feature is absent.
//   - Testability: Slim interfaces plus an unimplemented configuration reduce
//     boilerplate in fixtures and mock setups.
//
// Concurrency Guidance:
//   - Prefer immutable configuration structs after construction.
//   - Treat optional capabilities as independent; absence is not an error.
//   - Helper probes (Stdin, Stdout, Logger, Observability) never panic.
//
// To introduce a new capability (e.g., TracingAppContext):
//  1. Define an interface embedding AppContext[T] plus accessor(s).
//  2. Provide a generic helper: func Tracing[T ActionConfig](ctx AppContext[T]) (..., bool).
//  3. Implement the interface only in contexts needing that feature.
//
// Error Handling Philosophy: Capability helpers favor presence checks over
// sentinel errors, promoting resilient, feature-adaptive code paths without
// tightly coupling execution logic to wiring.
//
// Overall, this package encourages clear ownership of configuration, lean
// interfaces, and explicit optional feature discovery—yielding maintainable,
// testable application composition in Go. Package core defines the foundational
// abstractions for building CLI / service style applications around a
// strongly-typed, generic configuration, optional pipeline I/O, logging, and
// observability features.
//
// # What is a "configuration"
//
// In this model a "configuration" is any concrete type (struct) implementing
// the [ActionConfig] interface. It encapsulates all resources required by an
// application: log level, certificate material, secret source DSNs, tracing
// endpoint, metrics address, etc. By coding against the [ActionConfig]
// interface instead of a concrete struct, application logic can remain
// decoupled from the mechanism of loading / assembling configuration (flags,
// env, files, remote stores). Each application (or sub-command) is free to
// introduce its own richer configuration type while still satisfying the
// minimal contract.
//
// # Why use a generic (T ActionConfig)
//
// The generic parameter T on [AppContext] and related helpers provides:
//  1. Callers that obtain [AppContext.Config] receive the full concrete type
//     (T), not just the interface, eliminating repetitive casts.
//  2. Additional fields unique to a given application configuration are
//     immediately available wherever that specific T is in scope, with no need
//     for map access or reflection-based extraction.
//  3. Each executable unit (command/action) declares exactly which
//     configuration type it expects, improving readability and maintainability.
//  4. Go inlines interface method calls when feasible; generics avoid dynamic
//     type assertions at use sites for the concrete config.
//
// # Interfaces and Layering
//
// Rather than bloating AppContext with optional concerns, capabilities are
// modeled as additional interfaces that can be checked at runtime:
//
// [PipelineAppContext][T]:
//   - Stdin()/Stdout(): Declarative access to input/output streams enabling
//     pipeline-friendly commands without imposing streams on all contexts.
//
// [LoggerAppContext][T]:
//   - Log(): Provides a slog.Handler for structured logging emission without
//     forcing every context to carry a logger.
//
// [ObservabilityAppContext][T]:
//   - Observability(): Grants metrics instrumentation (Metrics interface) when
//     available; absent contexts remain lightweight.
//
// This list is not exhaustive, since runtime implementation technically MAY
// introduce some custom interfaces, however, it's highly recommended to keep the
// number of such interfaces minimal to reduce complexity.
//
// # Design Principles
//
//   - Separation of Concerns: Core logic depends only on interfaces; wiring layers
//     populate concrete implementations.
//   - Progressive Enhancement: Features (I/O, logging, metrics) are additive.
//   - Testability: UnimplementedActionConfig + small interfaces ease mock creation.
//   - Explicit Capability Discovery: Boolean return pattern communicates optionality.
//   - Type Safety via Generics: Reduces accidental mismatches and casts.
//
// # Extension Guidance
//
// To introduce new contextual capabilities (e.g., TracingAppContext, or
// whatever you want), define a new interface embedding [AppContext][T] plus
// accessor methods, and supply a helper similar to Logger() or Observability().
//
// By consolidating these patterns, the package offers a flexible, strongly
// typed, capability-driven foundation for building composable Go applications.
package core
