package logging

import (
	"context"
	"fmt"

	"golang.org/x/exp/slog"
)

type HandlerOption func(*slogHandler)

// HandlerWithCTXGroupName sets the namespace for data from the context.
// This can be used to prevent key collisions by nesting all data
// from the context in a Group.
func HandlerWithCTXGroupName(name string) HandlerOption {
	return func(sh *slogHandler) {
		sh.ctxGroupName = name
	}
}

// WrapHandler returns a handler that prints [ContextData],
// set with [Middleware] or custom additions.
//
// EXPERIMENTAL: API will break when we switch from `x/exp/slog` to `log/slog`
// when we drop Go <1.21 support.
func WrapHandler(h slog.Handler, opts ...HandlerOption) slog.Handler {
	// prevent wrapping if h is already a *slogHandler
	out, ok := h.(*slogHandler)
	if !ok {
		out = &slogHandler{
			handler: h,
		}
	}
	for _, opt := range opts {
		opt(out)
	}
	return out
}

// slogHandler implements the [slog.Handler] interface.
type slogHandler struct {
	handler      slog.Handler
	ctxGroupName string
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	handler := h.handler
	if data, ok := DataFromContext(ctx); ok {
		handler = handler.WithAttrs([]slog.Attr{
			slog.Any(h.ctxGroupName, data),
		})
	}
	return handler.Handle(ctx, record)
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{
		handler:      h.handler.WithAttrs(attrs),
		ctxGroupName: h.ctxGroupName,
	}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{
		handler:      h.handler.WithGroup(name),
		ctxGroupName: h.ctxGroupName,
	}
}

// StringValuer returns a slog.Valuer that
// forces the logger to use the type's String()
// method, even in json ouput mode.
// By wrapping the type we defer String()
// being called to the point we actually log.
func StringerValuer(s fmt.Stringer) slog.LogValuer {
	return stringerValuer{s}
}

type stringerValuer struct {
	fmt.Stringer
}

func (v stringerValuer) LogValue() slog.Value {
	return slog.StringValue(v.String())
}
