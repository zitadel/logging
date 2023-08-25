package logging

import (
	"context"
	"net/http"

	"github.com/benbjohnson/clock"
	"golang.org/x/exp/slog"
)

type ClientLoggerOption func(*logRountTripper)

func WithFallbackLogger(logger *slog.Logger) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.fallback = logger
	}
}

func WithClientClock(clock clock.Clock) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.clock = clock
	}
}

func WithClientRequestAttr(requestToAttr func(*http.Request) slog.Attr) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.reqToAttr = requestToAttr
	}
}

func WithClientResponseAttr(responseToAttr func(*http.Response) slog.Attr) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.resToAttr = responseToAttr
	}
}

// SetClientLogger adds slog functionality to the HTTP client.
// It attempts to obtain a logger with [FromContext].
// If no logger is in the context, it tries to use a fallback logger,
// which might be set by [WithFallbackLogger].
// If no logger was found finally, the Transport is
// executed without logging.
func SetClientLogger(c *http.Client, opts ...ClientLoggerOption) {
	lrt := &logRountTripper{
		next:      c.Transport,
		clock:     clock.New(),
		reqToAttr: requestToAttr,
		resToAttr: responseToAttr,
	}
	if lrt.next == nil {
		lrt.next = http.DefaultTransport
	}
	for _, opt := range opts {
		opt(lrt)
	}
	c.Transport = lrt
}

type logRountTripper struct {
	next     http.RoundTripper
	clock    clock.Clock
	fallback *slog.Logger

	reqToAttr func(*http.Request) slog.Attr
	resToAttr func(*http.Response) slog.Attr
}

// RoundTrip implements [http.RoundTripper].
func (l *logRountTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logger, ok := l.fromContextOrFallback(req.Context())
	if !ok {
		return l.next.RoundTrip(req)
	}
	start := l.clock.Now()

	resp, err := l.next.RoundTrip(req)
	logger = logger.With(
		l.reqToAttr(req),
		slog.Duration("duration", l.clock.Since(start)),
	)
	if err != nil {
		logger.Error("request roundtrip", "error", err)
		return resp, err
	}
	logger.Info("request roundtrip", l.resToAttr(resp))
	return resp, nil
}

func (l *logRountTripper) fromContextOrFallback(ctx context.Context) (*slog.Logger, bool) {
	if logger, ok := FromContext(ctx); ok {
		return logger, ok
	}
	return l.fallback, l.fallback != nil
}
