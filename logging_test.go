package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

var errTest = fmt.Errorf("im an error")
var callerFile = "github.com/zitadel/logging/logging_test.go"

func TestWithLogID(t *testing.T) {
	tests := []logTest{
		{
			"without error",
			Log("UTILS-B7l7"),
			map[string]interface{}{
				"logID": "UTILS-B7l7",
			},
		},
		{
			"with error",
			Log("UTILS-Ld9V").WithError(errTest),
			map[string]interface{}{
				"logID": "UTILS-Ld9V",
				"err":   errTest.Error(),
			},
		},
		{
			"on error",
			Log("UTILS-Ld9V").OnError(errTest),
			map[string]interface{}{
				"logID": "UTILS-Ld9V",
				"err":   errTest.Error(),
			},
		},
		{
			"on error without",
			Log("UTILS-Ld9V").OnError(nil),
			nil,
		},
		{
			"with fields",
			LogWithFields("LOGGI-5kk6z", "field1", 134, "field2", "asdlkfj"),
			map[string]interface{}{
				"logID":  "LOGGI-5kk6z",
				"field1": float64(134),
				"field2": "asdlkfj",
			},
		},
		{
			"with field",
			LogWithFields("LOGGI-5kk6z").WithField("field1", 134),
			map[string]interface{}{
				"logID":  "LOGGI-5kk6z",
				"field1": float64(134),
			},
		},
		{
			"fields odd",
			LogWithFields("LOGGI-xWzy4", "kevin"),
			map[string]interface{}{
				"logID":     "LOGGI-xWzy4",
				"oddFields": float64(1),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, testLogOutputFn(test))
	}
}

func TestWithoutLogID(t *testing.T) {
	tests := []logTest{
		{
			"without error",
			New(),
			map[string]interface{}{},
		},
		{
			"with error",
			New().WithError(errTest),
			map[string]interface{}{
				"err": errTest.Error(),
			},
		},
		{
			"on error",
			OnError(errTest),
			map[string]interface{}{
				"err": errTest.Error(),
			},
		},
		{
			"on error without",
			OnError(nil),
			nil,
		},
		{
			"with fields",
			WithFields("field1", 134, "field2", "asdlkfj"),
			map[string]interface{}{
				"field1": float64(134),
				"field2": "asdlkfj",
			},
		},
		{
			"with field",
			WithFields().WithField("field1", 134),
			map[string]interface{}{
				"field1": float64(134),
			},
		},
		{
			"fields odd",
			WithFields("kevin"),
			map[string]interface{}{
				"oddFields": float64(1),
			},
		},
		{
			"group attribute",
			New().WithField("group1", slog.GroupValue(slog.String("key", "value"))),
			map[string]interface{}{
				"group1": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, testLogOutputFn(test))
	}
}

type logTest struct {
	name           string
	entry          *Entry
	expectedOutput map[string]interface{}
}

func testLogOutputFn(test logTest) func(t *testing.T) {
	return func(t *testing.T) {
		var buf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelDebug,
		})))
		test.entry.Debug()
		bufBytes := buf.Bytes()
		if test.expectedOutput == nil {
			if len(bufBytes) > 0 {
				t.Fatalf("expected no output, got: %q", bufBytes)
			}
			return
		}
		var actual = make(map[string]interface{})
		if err := json.Unmarshal(bufBytes, &actual); err != nil {
			t.Fatalf("failed to unmarshal %q into map[string]interface{}: %v", bufBytes, err)
		}
		if _, err := time.Parse(time.RFC3339Nano, actual["time"].(string)); err != nil {
			t.Errorf("expected time in RFC3339Nano format, got: %q", actual["time"])
		}
		// We want to use reflect.DeepEqual later, so we remove the dynamic "time" field
		delete(actual, "time")
		test.expectedOutput["msg"] = ""
		test.expectedOutput["level"] = slog.LevelDebug.String()
		if !reflect.DeepEqual(actual, test.expectedOutput) {
			t.Errorf("expected output: %+v, got: %+v", test.expectedOutput, actual)
		}
	}
}
