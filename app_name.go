package core

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
