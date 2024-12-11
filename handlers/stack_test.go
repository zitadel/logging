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

var stackTestFile = "handlers/stack_test.go"

func TestStackHandler(t *testing.T) {
	tests := []struct {
		name                    string
		log                     func()
		expectedStackTraceStart string
		expectedOutput          map[string]interface{}
	}{
		{
			name: "Basic Handle",
			log:  func() { logging.Info("Test log message") },
			expectedOutput: map[string]interface{}{
				"msg":   "Test log message",
				"level": "INFO",
			},
		},
		{
			name: "WithAttrs adds attributes",
			log:  func() { logging.Info("Log with attributes", "key1", "value1", "key2", 42) },
			expectedOutput: map[string]interface{}{
				"msg":   "Log with attributes",
				"level": "INFO",
				"key1":  "value1",
				"key2":  float64(42), // Numbers will be unmarshalled as float64
			},
		},
		{
			name: "Info with GroupValue",
			log: func() {
				logging.Info("Log in group", "group1", slog.GroupValue(slog.String("key", "value")))
			},
			expectedOutput: map[string]interface{}{
				"msg":   "Log in group",
				"level": "INFO",
				"group1": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name: "WithFields with nested groups",
			log: func() {
				logging.WithFields("group1", slog.GroupValue(slog.Group("group2", slog.String("key", "value")))).Info("Log in nested group")
			},
			expectedOutput: map[string]interface{}{
				"msg":   "Log in nested group",
				"level": "INFO",
				"group1": map[string]interface{}{
					"group2": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
		{
			name: "LevelError adds stack trace",
			log:  func() { logging.OnError(errors.New("error message")).Error("an error happened") },
			expectedOutput: map[string]interface{}{
				"msg":   "an error happened",
				"level": "ERROR",
				"err":   "error message",
			},
			expectedStackTraceStart: "github.com/zitadel/logging/handlers_test.TestStackHandler",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			var buf bytes.Buffer
			slog.SetDefault(slog.New(handlers.AddCallerAndStack(slog.NewJSONHandler(&buf, nil))))
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
			if caller, ok := actual["caller"]; !ok || !strings.Contains(caller.(string), stackTestFile) {
				tt.Errorf("expected caller in %+v to contain %q, got: %q", actual, stackTestFile, caller)
			}
			// We want to use reflect.DeepEqual later, so we remove the dynamic "caller" field
			delete(actual, "caller")
			if !reflect.DeepEqual(actual, test.expectedOutput) {
				tt.Errorf("expected output: %+v, got: %+v", test.expectedOutput, actual)
			}
		})
	}
}
