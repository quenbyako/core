package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"
)

//nolint:ireturn // well, that's how env works
func parseURL(_ context.Context, v string) (any, error) {
	u, err := url.Parse(v)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	return *u, nil
}

//nolint:ireturn // well, that's how env works
func parseDuration(_ context.Context, v string) (any, error) {
	d, err := time.ParseDuration(v)
	if err != nil {
		return nil, fmt.Errorf("parse duration: %w", err)
	}

	return d, nil
}

//nolint:ireturn // well, that's how env works
func parseLocation(_ context.Context, v string) (any, error) {
	loc, err := time.LoadLocation(v)
	if err != nil {
		return nil, fmt.Errorf("parse location: %w", err)
	}

	return *loc, nil
}

//nolint:ireturn // well, that's how env works
func parseLogLevel(_ context.Context, input string) (any, error) {
	const (
		levelTrace = slog.LevelDebug - 4
		levelFatal = slog.LevelError + 4
		levelPanic = slog.LevelError + 8
	)

	switch strings.ToUpper(input) {
	case "TRACE":
		return levelTrace, nil
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	case "FATAL":
		return levelFatal, nil
	case "PANIC":
		return levelPanic, nil
	default:
		var l slog.Level
		if err := l.UnmarshalText([]byte(input)); err != nil {
			return nil, fmt.Errorf("parse log level: %w", err)
		}

		return l, nil
	}
}
