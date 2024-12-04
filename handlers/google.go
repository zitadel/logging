package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
)

// Constants for magic strings
const googleErrorType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"

func NewGoogle(w io.Writer, options *slog.HandlerOptions, data map[string]interface{}) slog.Handler {
	gWriter := &googleWriter{writer: w}
	var handler slog.Handler
	handler = slog.NewJSONHandler(gWriter, options)
	loggerAttributes := constructLoggerAttributes(data)
	if len(loggerAttributes) > 0 {
		handler = handler.WithAttrs(loggerAttributes)
	}
	return handler
}

// Helper to construct ServiceContext
func constructLoggerAttributes(data map[string]interface{}) []slog.Attr {
	if data == nil {
		return nil
	}
	if data["service"] == nil && data["version"] == nil {
		return nil
	}
	scValues := make([]any, 0, 2)
	if data["service"] != nil {
		scValues = append(scValues, slog.String("service", data["service"].(string)))
	}
	if data["version"] != nil {
		scValues = append(scValues, slog.String("version", data["version"].(string)))
	}
	return []slog.Attr{slog.Group("serviceContext", scValues...)}
}

type GoogleRecord struct {
	Time           string         `json:"time"`
	Message        string         `json:"message"`
	Severity       string         `json:"severity,omitempty"`
	Type           string         `json:"@type,omitempty"`
	StackTrace     string         `json:"stack_trace,omitempty"`
	AppContext     map[string]any `json:"appContext,omitempty"`
	ServiceContext map[string]any `json:"serviceContext,omitempty"`
}

func (c *googleWriter) mapAttributes(jsonHandlerOutput map[string]interface{}) *GoogleRecord {
	record := new(GoogleRecord)
	record.Message = jsonHandlerOutput["msg"].(string)
	record.Severity = jsonHandlerOutput["level"].(string)
	record.Time = jsonHandlerOutput["time"].(string)
	for key, value := range jsonHandlerOutput {
		switch key {
		case "level", "msg", "time":
		// Filter out
		case "err":
			record.Type = googleErrorType
			record.StackTrace = fmt.Sprintf("%s\n%s", record.Message, string(debug.Stack()))
			record.Message = fmt.Sprintf("%s: %s", record.Message, value.(string))
		case "serviceContext":
			record.ServiceContext = value.(map[string]any)
		default:
			if record.AppContext == nil {
				record.AppContext = make(map[string]any)
			}
			record.AppContext[key] = value
		}
	}
	return record
}

type googleWriter struct {
	writer io.Writer
}

func (g *googleWriter) Write(p []byte) (n int, err error) {
	// Unmarshal the JSON into a map
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return 0, fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	record := g.mapAttributes(logEntry)

	// Marshal the modified log entry back into JSON
	modifiedJSON, err := json.Marshal(record)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal modified log entry: %w", err)
	}

	// Write the modified JSON to the underlying writer
	return g.writer.Write(append(modifiedJSON, '\n'))
}
