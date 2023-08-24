package logging

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name string
		next func(*clock.Mock) http.HandlerFunc
		want string
	}{
		{
			name: "",
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
				"ctx":{
					"id":"id1",
					"request":{
						"method":"GET", "url":"https://example.com/path/"
					}
				},
				"response":{"duration":1000000000, "status":200, "written":13}
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOut := new(strings.Builder)
			logger := NewLogger(slog.NewJSONHandler(logOut, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}), "ctx").With("time", "not")

			clock := clock.NewMock()
			mw := Middleware(
				MiddlewareWithLoggerOption(logger),
				MiddlewareWithIDOption(func() string {
					return "id1"
				}),
				MiddlewareWithClockOptions(clock),
				MiddlewareWithGroupNames("request", "response"),
			)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "https://example.com/path/", nil)
			mw(tt.next(clock)).ServeHTTP(w, r)

			got := logOut.String()
			assert.JSONEq(t, tt.want, got)
		})
	}
}
