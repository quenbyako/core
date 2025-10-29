package core

import (
	"context"
)

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

const (
	// DefaultAppName is the fallback stable identifier used when no explicit
	// name is provided.
	DefaultAppName = "unknown"
	// DefaultAppTitle is the human-friendly title used when no explicit
	// application title is supplied.
	DefaultAppTitle = "Unknown Application"
)

// AppName represents both a machine-oriented stable identifier
// and a human-friendly display title for an application. Empty
// fields are normalized to defaults by [NewAppName] guaranteeing that
// calls to [AppName.Name]/[AppName.Title] always succeed with a non-empty
// fallback.
type AppName struct {
	name  string
	title string
}

// NewAppName constructs a new [AppName], normalizing empty inputs to
// [DefaultAppName]/[DefaultAppTitle]. Prefer passing explicit values when
// available; defaults keep logs and telemetry consistent for prototypes.
func NewAppName(name, title string) AppName {
	if name == "" {
		name = DefaultAppName
	}

	if title == "" {
		title = DefaultAppTitle
	}

	return AppName{
		name:  name,
		title: title,
	}
}

// Name returns the stable identifier and a boolean indicating whether it
// was explicitly set or derived from a default.
func (v AppName) Name() (string, bool) { return v.name, v.name != "" }

// Title returns the human-friendly application title with the same
// explicit/implicit semantics as Name.
func (v AppName) Title() (string, bool) { return v.title, v.title != "" }
