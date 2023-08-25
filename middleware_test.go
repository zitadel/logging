package logging

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func newTestLogger() (out *strings.Builder, logger *slog.Logger) {
	out = new(strings.Builder)
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}).WithAttrs([]slog.Attr{slog.String("time", "not")})
	return out, slog.New(handler)
}

type testWriter struct {
	*httptest.ResponseRecorder
	err error
}

func (w *testWriter) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	return w.ResponseRecorder.Write(b)
}

func newTestWriter(err error) *testWriter {
	return &testWriter{
		ResponseRecorder: httptest.NewRecorder(),
		err:              err,
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name string
		next func(*clock.Mock) http.HandlerFunc
		err  error
		want string
	}{
		{
			name: "ok",
			next: func(c *clock.Mock) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					c.Add(time.Second)
					fmt.Fprint(w, "Hello, World!")
				}
			},
			want: `{
				"level":"INFO",
				"time": "not",
				"msg":"request served",
				"id":"id1",
				"duration":1000000000,
				"request":{
					"method":"GET",
					"url":"https://example.com/path/"
				},
				"response":{
					"status":200,
					"written":13
				}
			}`,
		},
		{
			name: "error",
			next: func(c *clock.Mock) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					c.Add(time.Second)
					fmt.Fprint(w, "Hello, World!")
				}
			},
			err: io.ErrClosedPipe,
			want: `{
				"level":"WARN",
				"time": "not",
				"msg":"write response",
				"error": "io: read/write on closed pipe",
				"id":"id1",
				"duration":1000000000,
				"request":{
					"method":"GET",
					"url":"https://example.com/path/"
				},
				"response":{
					"status":200,
					"written":0
				}
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOut, logger := newTestLogger()

			clock := clock.NewMock()
			mw := Middleware(
				WithLogger(logger),
				WithIDFunc(func() slog.Attr {
					return slog.String("id", "id1")
				}),
				WithClock(clock),
				WithRequestAttr(requestToAttr),
				WithLoggedWriter(newLoggedWriter),
			)

			w := newTestWriter(tt.err)
			r := httptest.NewRequest("GET", "https://example.com/path/", nil)
			mw(tt.next(clock)).ServeHTTP(w, r)

			got := logOut.String()
			assert.JSONEq(t, tt.want, got)
		})
	}
}
