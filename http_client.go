package logging

import (
	"net/http"

	"github.com/benbjohnson/clock"
	"golang.org/x/exp/slog"
)

type ClientLoggerOption func(*logRountTripper)

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

func SetClientLogger(c *http.Client, logger *slog.Logger, opts ...ClientLoggerOption) {
	lrt := &logRountTripper{
		next:      c.Transport,
		clock:     clock.New(),
		logger:    logger,
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
	next   http.RoundTripper
	clock  clock.Clock
	logger *slog.Logger

	reqToAttr func(*http.Request) slog.Attr
	resToAttr func(*http.Response) slog.Attr
}

func (l *logRountTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := l.clock.Now()
	logger := l.logger.With(l.reqToAttr(req))

	resp, err := l.next.RoundTrip(req)
	logger = logger.With("duration", l.clock.Since(start))
	if err != nil {
		logger.Error("request roundtrip", "error", err)
		return resp, err
	}
	logger.Info("request roundtrip", l.resToAttr(resp))
	return resp, nil
}
