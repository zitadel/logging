package logging

import (
	"net/http"
	"net/url"

	"github.com/benbjohnson/clock"
	"golang.org/x/exp/slog"
)

type MiddlewareOption func(*middleware)

func MiddlewareWithLoggerOption(logger *slog.Logger) MiddlewareOption {
	return func(m *middleware) {
		m.logger = logger
	}
}

func MiddlewareWithIDOption(nextID func() string) MiddlewareOption {
	return func(m *middleware) {
		m.nextID = nextID
	}
}

func MiddlewareWithClockOptions(clock clock.Clock) MiddlewareOption {
	return func(m *middleware) {
		m.clock = clock
	}
}

func MiddlewareWithGroupNames(request, response string) MiddlewareOption {
	return func(m *middleware) {
		m.requestGroup = request
		m.responseGroup = response
	}
}

func Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mw := &middleware{
			logger:        WrapLogger(slog.Default(), "ctx"),
			requestGroup:  "request",
			responseGroup: "response",
			nextID:        func() string { return "" },
			clock:         clock.New(),
			next:          next,
		}
		for _, opt := range opts {
			opt(mw)
		}
		return mw
	}
}

type middleware struct {
	logger *slog.Logger
	nextID func() string
	next   http.Handler
	clock  clock.Clock

	requestGroup  string
	responseGroup string
}

func (m *middleware) setHTTPRequestData(r *http.Request, id string) *http.Request {
	data := NewContextData(id)
	data[m.requestGroup] = httpRequestData{
		method: r.Method,
		url:    *r.URL,
	}
	ctx := ContextWithData(r.Context(), data)
	return r.WithContext(ctx)
}

type httpRequestData struct {
	method string
	url    url.URL
}

// LogValue implements [slog.LogValuer]
func (d httpRequestData) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		slog.String("method", d.method),
		slog.Any("url", d.url.String()),
	}...)
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := m.clock.Now()

	r = m.setHTTPRequestData(r, m.nextID())
	lw := newLoggedWriter(w)
	m.next.ServeHTTP(lw, r)
	logger := m.logger.With(
		slog.Group(m.responseGroup, "duration", m.clock.Since(start), "status", lw.statusCode, "written", lw.written),
	)
	if lw.err != nil {
		logger.WarnContext(r.Context(), "response writer", "error", lw.err)
		return
	}
	logger.InfoContext(r.Context(), "request served")
}

type loggedWriter struct {
	http.ResponseWriter

	statusCode int
	written    int
	err        error
}

func newLoggedWriter(w http.ResponseWriter) *loggedWriter {
	return &loggedWriter{
		ResponseWriter: w,
	}
}

func (w *loggedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggedWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.written += n
	w.err = err
	return n, err
}
