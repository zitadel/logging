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
				"request":{
					"method":"GET",
					"url":"https://example.com/path/",
					"duration":1000000000,
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
				"request":{
					"method":"GET",
					"url":"https://example.com/path/",
					"duration":1000000000,
					"status":200,
					"written":0
				}
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOut := new(strings.Builder)
			handler := slog.NewJSONHandler(logOut, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}).WithAttrs([]slog.Attr{slog.String("time", "not")})
			logger := slog.New(WrapHandler(handler))

			clock := clock.NewMock()
			mw := Middleware(
				MiddlewareWithLoggerOption(logger),
				MiddlewareWithIDOption(func() string {
					return "id1"
				}),
				MiddlewareWithClockOptions(clock),
				MiddlewareWithGroupName("request"),
			)

			w := newTestWriter(tt.err)
			r := httptest.NewRequest("GET", "https://example.com/path/", nil)
			mw(tt.next(clock)).ServeHTTP(w, r)

			got := logOut.String()
			assert.JSONEq(t, tt.want, got)
		})
	}
}
