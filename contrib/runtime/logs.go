package runtime

import (
	"log/slog"
	"net"
)

const (
	eventEffectiveEnvironment = "notify.effective_environment"
)

type LogCallbacks interface {
	EffectiveEnvironment(env map[string]string)
	MetricsStarted(addr net.Addr)
	MetricsStopped(addr net.Addr)
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

func (l *logger) MetricsStarted(addr net.Addr) {
	l.log.Info(
		"Metrics server started",
		slog.Any("context", map[string]any{
			"addr": addr,
		}),
	)
}

func (l *logger) MetricsStopped(addr net.Addr) {
	l.log.Info(
		"Metrics server stopped",
		slog.Any("context", map[string]any{
			"addr": addr,
		}),
	)
}
