package logging

import (
	"net/http"

	"github.com/benbjohnson/clock"
	"golang.org/x/exp/slog"
)

type MiddlewareOption func(*middleware)

func MiddlewareWithHandlerOption(handler slog.Handler) MiddlewareOption {
	return func(m *middleware) {
		m.logger = slog.New(handler)
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

func MiddlewareWithGroupName(name string) MiddlewareOption {
	return func(m *middleware) {
		m.requestGroup = name
	}
}

func Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mw := &middleware{
			logger:       slog.New(WrapHandler(slog.Default().Handler())),
			requestGroup: "request",
			clock:        clock.New(),
			next:         next,
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

	requestGroup string
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := m.clock.Now()

	if m.nextID != nil {
		r = r.WithContext(
			ContextWithData(
				r.Context(),
				NewContextData(m.nextID()),
			),
		)
	}

	lw := newLoggedWriter(w)
	m.next.ServeHTTP(lw, r)
	logger := m.logger.With(
		slog.Group(m.requestGroup,
			"url", StringerValuer(r.URL),
			"method", r.Method,
			"duration", m.clock.Since(start),
			"status", lw.statusCode,
			"written", lw.written),
	)
	if lw.err != nil {
		logger.WarnContext(r.Context(), "write response", "error", lw.err)
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
