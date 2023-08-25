package logging

import (
	"context"

	"golang.org/x/exp/slog"
)

type ctxKeyType struct{}

var ctxKey ctxKeyType

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey, logger)
}
