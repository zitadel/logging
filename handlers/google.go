package handlers

import (
	"context"
	"fmt"
	"log/slog"
)

// Constants for magic strings
const googleErrorType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"

var _ slog.Handler = (*GoogleHandler)(nil)

type GoogleHandler struct {
	wrapped        slog.Handler
	serviceContext *slog.Attr
}

func ReplaceAttrForGoogleFunc(runAfterGoogleReplacements func([]string, slog.Attr) slog.Attr) func(groups []string, a slog.Attr) slog.Attr {
	after := func(groups []string, a slog.Attr) slog.Attr { return a }
	if runAfterGoogleReplacements != nil {
		after = runAfterGoogleReplacements
	}
	return func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case "level":
			a.Key = "severity"
		case "msg":
			a.Key = "message"
		}
		return after(groups, a)
	}
}

func ForGoogleCloudLogging(handler slog.Handler, configData map[string]any) *GoogleHandler {
	return &GoogleHandler{
		wrapped:        handler,
		serviceContext: constructLoggerAttributes(configData),
	}
}

// constructLoggerAttributes is a helper to construct ServiceContext
func constructLoggerAttributes(data map[string]interface{}) *slog.Attr {
	if data == nil {
		return nil
	}
	if data["service"] == nil && data["version"] == nil {
		return nil
	}
	var scAttr []any
	if service, ok := data["service"]; ok && service != "" {
		scAttr = append(scAttr, slog.String("service", service.(string)))
	}
	if version, ok := data["version"]; ok && version != "" {
		scAttr = append(scAttr, slog.String("version", version.(string)))
	}
	if len(scAttr) == 0 {
		return nil
	}
	sc := slog.Group("service_context", scAttr...)
	return &sc
}

func (g *GoogleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return g.wrapped.Enabled(ctx, level)
}

func (g *GoogleHandler) Handle(ctx context.Context, record slog.Record) error {
	newAttrs := make([]slog.Attr, 0, 2)
	if g.serviceContext != nil {
		newAttrs = append(newAttrs, *g.serviceContext)
	}
	if record.Level >= slog.LevelError {
		newAttrs = append(newAttrs, slog.String("@type", googleErrorType))
	}
	appContextAttrs := make([]any, 0, record.NumAttrs())
	originalMessage := record.Message
	record.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "err", "error":
			if record.Level >= slog.LevelError {
				record.Message = fmt.Sprintf("%s: %s", record.Message, a.Value)
				return true
			}
		case "stack_trace":
			// Stays top level
			newAttrs = append(newAttrs, slog.String("stack_trace", fmt.Sprintf("%s\n%s", originalMessage, a.Value)))
			return true
		}
		appContextAttrs = append(appContextAttrs, a)
		return true
	})
	if len(appContextAttrs) > 0 {
		newAttrs = append(newAttrs, slog.Group("app_context", appContextAttrs...))
	}
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(newAttrs...)
	return g.wrapped.Handle(ctx, newRecord)
}

func (g *GoogleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	g.wrapped = g.wrapped.WithAttrs(attrs)
	return g
}

func (g *GoogleHandler) WithGroup(name string) slog.Handler {
	g.wrapped = g.wrapped.WithGroup(name)
	return g
}
