package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/zitadel/logging"
	"github.com/zitadel/logging/handlers"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"testing"
)

var callerFile = "handlers/google_test.go"

func TestGoogleHandler(t *testing.T) {
	tests := []struct {
		name                    string
		handler                 func(writer io.Writer) slog.Handler
		log                     func()
		expectedOutput          map[string]interface{}
		expectedStackTraceStart string
	}{
		{
			name:    "Basic Handle",
			handler: func(writer io.Writer) slog.Handler { return handlers.NewGoogle(writer, nil, nil) },
			log:     func() { logging.Info("Test log message") },
			expectedOutput: map[string]interface{}{
				"message":  "Test log message",
				"severity": "INFO",
			},
		},
		{
			name: "WithAttrs adds attributes",
			handler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, nil).WithAttrs([]slog.Attr{
					slog.String("key1", "value1"),
					slog.Int("key2", 42),
				})
			},
			log: func() { logging.Info("Log with attributes") },
			expectedOutput: map[string]interface{}{
				"message":  "Log with attributes",
				"severity": "INFO",
				"app_context": map[string]interface{}{
					"key1": "value1",
					"key2": float64(42), // Numbers will be unmarshaled as float64
				},
			},
		},
		{
			name:    "WithGroup groups attributes",
			handler: func(writer io.Writer) slog.Handler { return handlers.NewGoogle(writer, nil, nil).WithGroup("group1") },
			log:     func() { logging.Info("Log in group") },
			expectedOutput: map[string]interface{}{
				"message":  "Log in group",
				"severity": "INFO",
				"app_context": map[string]interface{}{
					"group1": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
		{
			name: "WithGroup nested groups",
			handler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, nil).WithGroup("group1").WithGroup("group2")
			},
			log: func() { logging.Info("Log in nested group") },
			expectedOutput: map[string]interface{}{
				"message":  "Log in nested group",
				"severity": "INFO",
				"app_context": map[string]interface{}{
					"group1": map[string]interface{}{
						"group2": map[string]interface{}{
							"key": "value",
						},
					},
				},
			},
		},
		{
			name:    "LevelError adds error type and stack trace",
			handler: func(writer io.Writer) slog.Handler { return handlers.NewGoogle(writer, nil, nil) },
			log:     func() { logging.OnError(errors.New("error message")).Error("an error happened") },
			expectedOutput: map[string]interface{}{
				"message":  "an error happened: error message",
				"severity": "ERROR",
				"@type":    "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
			},
			expectedStackTraceStart: "an error happened\ngithub.com/zitadel/logging/handlers_test.TestGoogleHandler",
		},
		{
			name: "Service and version are added to the service context group",
			handler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, map[string]interface{}{
					"service": "test-service",
					"version": "1.0.0",
				})
			},
			log: func() { logging.Info("Log with service context") },
			expectedOutput: map[string]interface{}{
				"message":  "Log with service context",
				"severity": "INFO",
				"service_context": map[string]interface{}{
					"service": "test-service",
					"version": "1.0.0",
				},
			},
		},
		{
			name:    "err field on info level is unchanged",
			handler: func(writer io.Writer) slog.Handler { return handlers.NewGoogle(writer, nil, nil) },
			log:     func() { logging.Info("Info log with err field", "err", errors.New("error message")) },
			expectedOutput: map[string]interface{}{
				"message":  "Info log with err field",
				"severity": "INFO",
				"app_context": map[string]interface{}{
					"err": "error message",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			var buf bytes.Buffer
			slog.SetDefault(slog.New(test.handler(&buf)))
			test.log()
			var actual map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &actual); err != nil {
				tt.Fatalf("failed to unmarshal log: %v", err)
			}
			if logtime, ok := actual["time"]; !ok || logtime == "" {
				tt.Errorf("expected time field, got: %q", logtime)
			}
			// We want to use reflect.DeepEqual later, so we remove the dynamic "time" field
			delete(actual, "time")
			if test.expectedStackTraceStart != "" {
				if stackTrace, ok := actual["stack_trace"]; !ok || !strings.HasPrefix(stackTrace.(string), test.expectedStackTraceStart) {
					tt.Errorf("expected stack trace in %+v to start with %q, got: %q", actual, test.expectedStackTraceStart, stackTrace)
				}
			}
			// We want to use reflect.DeepEqual later, so we remove the dynamic "stack_trace" field
			delete(actual, "stack_trace")
			if caller, ok := actual["app_context"].(map[string]interface{})["caller"]; !ok || !strings.Contains(caller.(string), callerFile) {
				tt.Errorf("expected caller in %+v to contain %q, got: %q", actual, callerFile, caller)
			}
			// We want to use reflect.DeepEqual later, so we remove the dynamic "caller" field
			delete(actual["app_context"].(map[string]interface{}), "caller")
			// We delete an empty appContext so we don't have to expect
			if appContext, ok := actual["app_context"]; ok && len(appContext.(map[string]interface{})) == 0 {
				delete(actual, "app_context")
			}
			if !reflect.DeepEqual(actual, test.expectedOutput) {
				tt.Errorf("expected output: %+v, got: %+v", test.expectedOutput, actual)
			}
		})
	}
}
