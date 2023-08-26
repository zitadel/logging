package logging

import (
	"net/http"

	"github.com/benbjohnson/clock"
	"golang.org/x/exp/slog"
)

type MiddlewareOption func(*middleware)

func WithLogger(logger *slog.Logger) MiddlewareOption {
	return func(m *middleware) {
		m.logger = logger
	}
}

// WithGroup groups the log attributes
// produced by the middleware.
func WithGroup(name string) MiddlewareOption {
	return func(m *middleware) {
		m.group = name
	}
}

func WithIDFunc(nextID func() slog.Attr) MiddlewareOption {
	return func(m *middleware) {
		m.nextID = nextID
	}
}

func WithClock(clock clock.Clock) MiddlewareOption {
	return func(m *middleware) {
		m.clock = clock
	}
}

func WithRequestAttr(requestToAttr func(*http.Request) slog.Attr) MiddlewareOption {
	return func(m *middleware) {
		m.reqAttr = requestToAttr
	}
}

func WithLoggedWriter(wrap func(w http.ResponseWriter) LoggedWriter) MiddlewareOption {
	return func(m *middleware) {
		m.wrapWriter = wrap
	}
}

func Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mw := &middleware{
			logger:     slog.Default(),
			clock:      clock.New(),
			next:       next,
			reqAttr:    requestToAttr,
			wrapWriter: newLoggedWriter,
		}
		for _, opt := range opts {
			opt(mw)
		}
		return mw
	}
}

type middleware struct {
	logger     *slog.Logger
	group      string
	nextID     func() slog.Attr
	next       http.Handler
	clock      clock.Clock
	reqAttr    func(*http.Request) slog.Attr
	wrapWriter func(http.ResponseWriter) LoggedWriter
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := m.clock.Now()

	logger := m.logger.With(slog.Group(m.group, m.reqAttr(r)))
	if m.nextID != nil {
		logger = logger.With(slog.Group(m.group, m.nextID()))
	}
	r = r.WithContext(ToContext(r.Context(), logger))

	lw := m.wrapWriter(w)
	m.next.ServeHTTP(lw, r)
	logger = logger.With(slog.Group(m.group,
		slog.Duration("duration", m.clock.Since(start)),
		lw.Attr(),
	))
	if err := lw.Err(); err != nil {
		logger.WarnContext(r.Context(), "write response", "error", err)
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

func newLoggedWriter(w http.ResponseWriter) LoggedWriter {
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

func (lw *loggedWriter) Attr() slog.Attr {
	return slog.Group("response",
		"status", lw.statusCode,
		"written", lw.written,
	)
}

func (lw *loggedWriter) Err() error {
	return lw.err
}
