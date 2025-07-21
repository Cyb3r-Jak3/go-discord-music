package bot

import (
	"context"
	"log/slog"

	"github.com/sirupsen/logrus"
)

// LogrusAdapter adapts logrus to slog.Handler so it can be used with github.com/disgoorg/disgolink.
type LogrusAdapter struct {
	logger *logrus.Logger
}

func NewLogrusAdapter(logger *logrus.Logger) *LogrusAdapter {
	return &LogrusAdapter{logger: logger}
}

func (a *LogrusAdapter) Handle(_ context.Context, r slog.Record) error {
	attrs := make(logrus.Fields)
	r.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value
		return true
	})
	attrs["service"] = "disgolink"
	entry := a.logger.WithFields(attrs)
	switch r.Level {
	case slog.LevelDebug:
		entry.Debug(r.Message)
	case slog.LevelInfo:
		entry.Info(r.Message)
	case slog.LevelWarn:
		entry.Warn(r.Message)
	case slog.LevelError:
		entry.Error(r.Message)
	default:
		entry.Info(r.Message)
	}
	return nil
}

func (a *LogrusAdapter) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (a *LogrusAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make(logrus.Fields)
	for _, attr := range attrs {
		fields[attr.Key] = attr.Value
	}
	return NewLogrusAdapter(a.logger)
}

func (a *LogrusAdapter) WithGroup(_ string) slog.Handler {
	return a
}
