package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/zitadel/logging"
	"github.com/zitadel/logging/handlers"
	"log/slog"
	"reflect"
	"strings"
	"testing"
)

var callerFile = "handlers/google_test.go"

func TestGoogleHandler(t *testing.T) {
	tests := []struct {
		name                    string
		config                  map[string]any
		log                     func()
		expectedStackTraceStart string
		expectedOutput          map[string]interface{}
	}{
		{
			name: "Basic Handle",
			log:  func() { logging.Info("Test log message") },
			expectedOutput: map[string]interface{}{
				"message":  "Test log message",
				"severity": "INFO",
			},
		},
		{
			name: "WithAttrs adds attributes",
			log:  func() { logging.Info("Log with attributes", "key1", "value1", "key2", 42) },
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
			name: "Info with GroupValue",
			log: func() {
				logging.Info("Log in group", "group1", slog.GroupValue(slog.String("key", "value")))
			},
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
			name: "WithFields with nested groups",
			log: func() {
				logging.WithFields("group1", slog.GroupValue(slog.Group("group2", slog.String("key", "value")))).Info("Log in nested group")
			},
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
			name: "LevelError adds error type and stack trace",
			log:  func() { logging.OnError(errors.New("error message")).Error("an error happened") },
			expectedOutput: map[string]interface{}{
				"message":  "an error happened: error message",
				"severity": "ERROR",
				"@type":    "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
			},
			expectedStackTraceStart: "an error happened\ngithub.com/zitadel/logging/handlers_test.TestGoogleHandler",
		},
		{
			name: "Service and version are added to the service context group",
			config: map[string]any{
				"service": "test-service",
				"version": "1.0.0",
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
			name: "err field on info level is unchanged",
			log: func() {
				logging.Info("Info log with err field", "err", errors.New("error message"))
			},
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
			var handler slog.Handler
			handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				ReplaceAttr: handlers.ReplaceAttrForGoogleFunc(nil),
			})
			handler = handlers.ForGoogleCloudLogging(handler, test.config)
			if test.expectedStackTraceStart != "" {
				handler = handlers.AddCallerAndStack(handler)
			}
			slog.SetDefault(slog.New(handler))
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
				// We want to use reflect.DeepEqual later, so we remove the dynamic "stack_trace" field
				delete(actual, "stack_trace")
				appContext, ok := actual["app_context"]
				if !ok {
					tt.Fatalf("expected app_context in %+v, got none", actual)
				}
				if caller, ok := appContext.(map[string]interface{})["caller"]; !ok || !strings.Contains(caller.(string), callerFile) {
					tt.Errorf("expected caller in %+v to contain %q, got: %q", actual, callerFile, caller)
				}
				// We want to use reflect.DeepEqual later, so we remove the dynamic "caller" field
				delete(actual["app_context"].(map[string]interface{}), "caller")
			}
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
