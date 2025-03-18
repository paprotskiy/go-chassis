package logger

import (
	"context"
	"io"
	"log/slog"
)

const verName = "version"

func GetLogLevel(isDebug bool) slog.Level {
	return map[bool]slog.Level{
		true:  slog.LevelDebug,
		false: slog.LevelInfo,
	}[isDebug]
}

func NewJsonLogger(ctx context.Context, w io.Writer, logLevel slog.Level, version string) *Logger {
	level := new(slog.LevelVar)
	level.Set(logLevel)

	return &Logger{
		ctx: ctx,
		ver: []any{verName, version},
		sl: slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: level,
		})),
	}
}

type Logger struct {
	ctx context.Context
	ver []any
	sl  *slog.Logger
}

func (l *Logger) buildArgs(args []any) []any {
	return append(l.ver, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.sl.DebugContext(l.ctx, msg, l.buildArgs(args)...)
}

func (l *Logger) Error(err error) {
	l.sl.ErrorContext(l.ctx, err.Error())
}

func (l *Logger) Info(msg string, args ...any) {
	l.sl.InfoContext(l.ctx, msg, l.buildArgs(args)...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.sl.WarnContext(l.ctx, msg, l.buildArgs(args)...)
}
