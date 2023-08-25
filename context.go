package logging

import (
	"context"

	"golang.org/x/exp/slog"
)

type ctxKeyType struct{}

var ctxKey ctxKeyType

func FromContext(ctx context.Context) (logger *slog.Logger, ok bool) {
	logger, ok = ctx.Value(ctxKey).(*slog.Logger)
	return logger, ok
}

func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey, logger)
}
