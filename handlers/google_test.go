package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/zitadel/logging/handlers"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGoogleHandler(t *testing.T) {
	logTime := time.Now()
	tests := []struct {
		name                     string
		setupHandler             func(writer io.Writer) slog.Handler
		record                   slog.Record
		expectedOutput           map[string]interface{}
		expectedStackStraceStart string
	}{
		{
			name:         "Basic Handle",
			setupHandler: func(writer io.Writer) slog.Handler { return handlers.NewGoogle(writer, nil, nil) },
			record: slog.Record{
				Time:    logTime,
				Level:   slog.LevelInfo,
				Message: "Test log message",
			},
			expectedOutput: map[string]interface{}{
				"time":     logTime.Format(time.RFC3339Nano), // Update with dynamic time
				"message":  "Test log message",
				"severity": "INFO",
			},
		},
		{
			name: "WithAttrs adds attributes",
			setupHandler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, nil).WithAttrs([]slog.Attr{
					slog.String("key1", "value1"),
					slog.Int("key2", 42),
				})
			},
			record: slog.Record{
				Time:    logTime,
				Level:   slog.LevelInfo,
				Message: "Log with attributes",
			},
			expectedOutput: map[string]interface{}{
				"time":     logTime.Format(time.RFC3339Nano),
				"message":  "Log with attributes",
				"severity": "INFO",
				"appContext": map[string]interface{}{
					"key1": "value1",
					"key2": float64(42), // Numbers will be unmarshaled as float64
				},
			},
		},
		{
			name: "WithGroup groups attributes",
			setupHandler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, nil).WithGroup("group1").WithAttrs([]slog.Attr{
					slog.String("key", "value"),
				})
			},
			record: slog.Record{
				Time:    logTime,
				Level:   slog.LevelInfo,
				Message: "Log in group",
			},
			expectedOutput: map[string]interface{}{
				"time":     logTime.Format(time.RFC3339Nano),
				"message":  "Log in group",
				"severity": "INFO",
				"appContext": map[string]interface{}{
					"group1": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
		{
			name: "WithGroup nested groups",
			setupHandler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, nil).WithGroup("group1").WithGroup("group2").WithAttrs([]slog.Attr{
					slog.String("key", "value"),
				})
			},
			record: slog.Record{
				Time:    logTime,
				Level:   slog.LevelInfo,
				Message: "Log in nested group",
			},
			expectedOutput: map[string]interface{}{
				"time":     logTime.Format(time.RFC3339Nano),
				"message":  "Log in nested group",
				"severity": "INFO",
				"appContext": map[string]interface{}{
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
			setupHandler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, nil).WithAttrs([]slog.Attr{
					slog.Any("err", errors.New("error message")),
				})
			},
			record: slog.Record{
				Time:    logTime,
				Level:   slog.LevelError,
				Message: "an error happened",
			},
			expectedOutput: map[string]interface{}{
				"time":     logTime.Format(time.RFC3339Nano),
				"message":  "an error happened: error message",
				"severity": "ERROR",
				"@type":    "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
			},
			expectedStackStraceStart: "an error happened\ngoroutine",
		},
		{
			name: "Service and version are added to the service context group",
			setupHandler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, map[string]interface{}{
					"service": "test-service",
					"version": "1.0.0",
				})
			},
			record: slog.Record{
				Time:    logTime,
				Level:   slog.LevelInfo,
				Message: "Log with service context",
			},
			expectedOutput: map[string]interface{}{
				"time":     logTime.Format(time.RFC3339Nano),
				"message":  "Log with service context",
				"severity": "INFO",
				"serviceContext": map[string]interface{}{
					"service": "test-service",
					"version": "1.0.0",
				},
			},
		},
		{
			name: "Service and version are not added to the service context group if missing",
			setupHandler: func(writer io.Writer) slog.Handler {
				return handlers.NewGoogle(writer, nil, map[string]interface{}{})
			},
			record: slog.Record{
				Time:    logTime,
				Level:   slog.LevelInfo,
				Message: "Log without service context",
			},
			expectedOutput: map[string]interface{}{
				"time":     logTime.Format(time.RFC3339Nano),
				"message":  "Log without service context",
				"severity": "INFO",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := tt.setupHandler(&buf)
			err := handler.Handle(context.Background(), tt.record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var actual map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &actual); err != nil {
				t.Fatalf("failed to unmarshal log: %v", err)
			}
			if tt.expectedStackStraceStart != "" {
				if !strings.HasPrefix(actual["stack_trace"].(string), tt.expectedStackStraceStart) {
					t.Errorf("expected stack trace to start with %q, got: %q", tt.expectedStackStraceStart, actual["stack_trace"])
				}
				// Remove the stack trace from the actual output, so we can use reflect.DeepEqual later
				delete(actual, "stack_trace")
			}
			if !reflect.DeepEqual(actual, tt.expectedOutput) {
				t.Errorf("expected output: %+v, got: %+v", tt.expectedOutput, actual)
			}
		})
	}
}
