package runtime

import "log/slog"

const (
	eventEffectiveEnvironment = "notify.effective_environment"
)

type LogCallbacks interface {
	EffectiveEnvironment(env map[string]string)
}

type logger struct {
	log *slog.Logger
}

var _ LogCallbacks = (*logger)(nil)

func defaultLogs(l slog.Handler) LogCallbacks {
	return &logger{log: slog.New(l)}
}

func (l *logger) EffectiveEnvironment(env map[string]string) {
	l.log.Info(
		"Parsed effective environment",
		slog.String("event_type", eventEffectiveEnvironment),
		slog.Any("context",
			map[string]any{
				"env": env,
			},
		),
	)
}
