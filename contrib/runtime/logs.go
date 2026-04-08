package runtime

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
)

const (
	eventEffectiveEnvironment = "notify.effective_environment"
)

type LogCallbacks interface {
	EffectiveEnvironment(ctx context.Context, env map[string]string)
	MetricsStarted(ctx context.Context, addr net.Addr)
	MetricsStopped(ctx context.Context, addr net.Addr)
}

type logger struct {
	log log.Logger
}

var _ LogCallbacks = (*logger)(nil)

func defaultLogs(l log.LoggerProvider) LogCallbacks {
	return &logger{log: l.Logger("runtime")}
}

func (l *logger) EffectiveEnvironment(ctx context.Context, env map[string]string) {
	var record log.Record
	record.SetSeverity(log.SeverityInfo)
	record.SetEventName(eventEffectiveEnvironment)
	record.SetBody(log.StringValue("Parsed effective environment"))

	record.AddAttributes(
		log.KeyValueFromAttribute(attribute.Key("env").String(fmt.Sprintln(env))),
	)

	l.log.Emit(ctx, record)
}

func (l *logger) MetricsStarted(ctx context.Context, addr net.Addr) {
	var record log.Record
	record.SetSeverity(log.SeverityInfo)
	record.SetEventName("notify.metrics_started")
	record.SetBody(log.StringValue("Metrics server started"))
	record.AddAttributes(
		log.KeyValueFromAttribute(attribute.Key("addr").String(addr.String())),
	)

	l.log.Emit(ctx, record)
}

func (l *logger) MetricsStopped(ctx context.Context, addr net.Addr) {
	var record log.Record
	record.SetSeverity(log.SeverityInfo)
	record.SetEventName("notify.metrics_stopped")
	record.SetBody(log.StringValue("Metrics server stopped"))
	record.AddAttributes(
		log.KeyValueFromAttribute(attribute.Key("addr").String(addr.String())),
	)

	l.log.Emit(ctx, record)
}
