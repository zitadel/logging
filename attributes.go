package logging

import (
	"fmt"
	"net/http"

	"golang.org/x/exp/slog"
)

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

func requestToAttr(req *http.Request) slog.Attr {
	return slog.Group("request",
		slog.String("method", req.Method),
		slog.Any("url", StringerValuer(req.URL)),
	)
}

func responseToAttr(resp *http.Response) slog.Attr {
	return slog.Group("response",
		slog.String("status", resp.Status),
		slog.Int64("content_length", resp.ContentLength),
	)
}

type LoggedWriter interface {
	http.ResponseWriter
	Attr() slog.Attr
	Err() error
}
