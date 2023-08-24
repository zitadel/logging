package logging

import (
	"context"

	"golang.org/x/exp/slog"
)

// WrapLogger configures the logger to print [ContextData],
// set with [Middleware] or custom additions.
//
// If logger is nil, [slog.Default] will be used.
//
// ctxDataGroup can be set to namespace data from the context.
// If empty, the Attributes generated from the context will be
// inlined with the other fields in the Logger and care should be taken
// to prevent name collisions.
//
// EXPERIMENTAL: API will break when we switch from `x/exp/slog` to `log/slog`
// when we drop Go <1.21 support.
func WrapLogger(logger *slog.Logger, ctxDataGroup string) *slog.Logger {
	if logger == nil {
		logger = slog.Default()
	}
	return slog.New(&slogHandler{
		handler:      logger.Handler(),
		ctxDataGroup: ctxDataGroup,
	})
}

// NewLogger creates a new logger that prints [ContextData],
// set with [Middleware] or custom additions.
//
// ctxDataGroup can be set to namespace data from the context.
// If empty, the Attributes generated from the context will be
// inlined with the other fields in the Logger and care should be taken
// to prevent name collisions.
//
// EXPERIMENTAL: API will break when we switch from `x/exp/slog` to `log/slog`
// when we drop Go <1.21 support.
func NewLogger(h slog.Handler, ctxDataGroup string) *slog.Logger {
	if h == nil {
		panic("nil Handler")
	}
	return slog.New(
		&slogHandler{
			handler:      h,
			ctxDataGroup: ctxDataGroup,
		},
	)
}

// slogHandler implements the [slog.Handler] interface.
type slogHandler struct {
	handler      slog.Handler
	ctxDataGroup string
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	handler := h.handler
	if data, ok := DataFromContext(ctx); ok {
		handler = handler.WithAttrs([]slog.Attr{
			slog.Any(h.ctxDataGroup, data),
		})
	}
	return handler.Handle(ctx, record)
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{
		handler:      h.handler.WithAttrs(attrs),
		ctxDataGroup: h.ctxDataGroup,
	}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{
		handler:      h.handler.WithGroup(name),
		ctxDataGroup: h.ctxDataGroup,
	}
}
