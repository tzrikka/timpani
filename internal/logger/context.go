// Package logger provides utilities for working with [slog] and [context.Context].
package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type ctxKey struct{}

var ctxLoggerKey = ctxKey{}

func InContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if ctxLogger, ok := ctx.Value(ctxLoggerKey).(*slog.Logger); ok {
		l = ctxLogger
	}
	return l
}

func Fatal(ctx context.Context, msg string, attrs ...slog.Attr) {
	fatalErrorCtx(ctx, msg, nil, attrs...)
}

func FatalError(msg string, err error, attrs ...slog.Attr) {
	fatalErrorCtx(context.Background(), msg, err, attrs...)
}

func FatalErrorContext(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	fatalErrorCtx(ctx, msg, err, attrs...)
}

func fatalErrorCtx(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // Discard wrapper frames (Callers, fatalErrorCtx, Fatal*).

	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
	if err != nil {
		r.AddAttrs(slog.Any("error", err))
	}
	r.AddAttrs(attrs...)

	_ = slog.Default().Handler().Handle(ctx, r)
	os.Exit(1)
}
