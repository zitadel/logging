package logging

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errRountripper struct{}

func (errRountripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, io.ErrClosedPipe
}

func Test_EnableHTTPClient(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		fromCtx   bool
		wantErr   error
		wantLog   string
	}{
		{
			name:      "nil transport / default",
			transport: nil,
			wantLog: `{
				"level":"INFO",
				"msg":"request roundtrip",
				"time":"not",
				"request":{"method":"GET","url":"%s"},
				"duration":1000000000,
				"response":{
					"status":"200 OK",
					"content_length":14
				}
			}`,
		},
		{
			name:      "transport set",
			transport: http.DefaultTransport,
			wantLog: `{
				"level":"INFO",
				"msg":"request roundtrip",
				"time":"not",
				"request":{"method":"GET","url":"%s"},
				"duration":1000000000,
				"response":{
					"status":"200 OK",
					"content_length":14
				}
			}`,
		},
		{
			name:      "roundtrip error",
			transport: errRountripper{},
			wantErr:   io.ErrClosedPipe,
			wantLog: `{
				"level":"ERROR",
				"msg":"request roundtrip",
				"time":"not",
				"request":{"method":"GET","url":"%s"},
				"error":"io: read/write on closed pipe",
				"duration":0
			}`,
		},
		{
			name:      "logger from ctx",
			transport: http.DefaultTransport,
			fromCtx:   true,
			wantLog: `{
				"level":"INFO",
				"msg":"request roundtrip",
				"time":"not",
				"ctx":{
					"request":{"method":"GET","url":"%s"},
					"duration":1000000000,
					"response":{
						"status":"200 OK",
						"content_length":14
					}
				}
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, logger := newTestLogger()
			clock := clock.NewMock()
			c := &http.Client{
				Transport: tt.transport,
			}
			EnableHTTPClient(c,
				WithFallbackLogger(logger),
				WithClientClock(clock),
				WithClientRequestAttr(requestToAttr),
				WithClientResponseAttr(responseToAttr),
			)

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello, client")
				clock.Add(time.Second)
			}))
			defer ts.Close()

			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			require.NoError(t, err)

			if tt.fromCtx {
				req = req.WithContext(ToContext(req.Context(), logger.WithGroup("ctx")))
			}
			_, err = c.Do(req)
			require.ErrorIs(t, err, tt.wantErr)

			wantLog := fmt.Sprintf(tt.wantLog, ts.URL)
			assert.JSONEq(t, wantLog, out.String())
		})
	}
}
