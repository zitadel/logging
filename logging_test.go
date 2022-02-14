package logging

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

var errTest = fmt.Errorf("im an error")

func TestWithLogID(t *testing.T) {
	tests := []struct {
		name           string
		entry          *Entry
		expectedFields map[string]func(interface{}) bool
	}{
		{
			"without error",
			Log("UTILS-B7l7"),
			map[string]func(interface{}) bool{
				"logID": func(got interface{}) bool {
					logID, ok := got.(string)
					if !ok {
						return false
					}
					return logID == "UTILS-B7l7"
				},
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
			},
		},
		{
			"with error",
			Log("UTILS-Ld9V").WithError(errTest),
			map[string]func(interface{}) bool{
				"logID": func(got interface{}) bool {
					logID, ok := got.(string)
					if !ok {
						return false
					}
					return logID == "UTILS-Ld9V"
				},
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"error": func(got interface{}) bool {
					err, ok := got.(error)
					if !ok {
						return false
					}
					return errors.Is(err, errTest)
				},
			},
		},
		{
			"on error",
			Log("UTILS-Ld9V").OnError(errTest),
			map[string]func(interface{}) bool{
				"logID": func(got interface{}) bool {
					logID, ok := got.(string)
					if !ok {
						return false
					}
					return logID == "UTILS-Ld9V"
				},
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"error": func(got interface{}) bool {
					err, ok := got.(error)
					if !ok {
						return false
					}
					return errors.Is(err, errTest)
				},
			},
		},
		{
			"on error without",
			Log("UTILS-Ld9V").OnError(nil),
			map[string]func(interface{}) bool{
				"logID": func(got interface{}) bool {
					logID, ok := got.(string)
					if !ok {
						return false
					}
					return logID == "UTILS-Ld9V"
				},
			},
		},
		{
			"with fields",
			LogWithFields("LOGGI-5kk6z", "field1", 134, "field2", "asdlkfj"),
			map[string]func(interface{}) bool{
				"logID": func(got interface{}) bool {
					logID, ok := got.(string)
					if !ok {
						return false
					}
					return logID == "LOGGI-5kk6z"
				},
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"field1": func(got interface{}) bool {
					i, ok := got.(int)
					if !ok {
						return false
					}
					return i == 134
				},
				"field2": func(got interface{}) bool {
					i, ok := got.(string)
					if !ok {
						return false
					}
					return i == "asdlkfj"
				},
			},
		},
		{
			"with field",
			LogWithFields("LOGGI-5kk6z").WithField("field1", 134),
			map[string]func(interface{}) bool{
				"logID": func(got interface{}) bool {
					logID, ok := got.(string)
					if !ok {
						return false
					}
					return logID == "LOGGI-5kk6z"
				},
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"field1": func(got interface{}) bool {
					i, ok := got.(int)
					if !ok {
						return false
					}
					return i == 134
				},
			},
		},
		{
			"fields odd",
			LogWithFields("LOGGI-xWzy4", "kevin"),
			map[string]func(interface{}) bool{
				"logID": func(got interface{}) bool {
					logID, ok := got.(string)
					if !ok {
						return false
					}
					return logID == "LOGGI-xWzy4"
				},
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"oddFields": func(got interface{}) bool {
					i, ok := got.(int)
					if !ok {
						return false
					}
					return i == 1
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.entry.Debug()
			if len(test.entry.Data) != len(test.expectedFields) {
				t.Errorf("enexpected amount of fields got: %d, want %d", len(test.entry.Data), len(test.expectedFields))
			}
			for key, expectedValue := range test.expectedFields {
				value, ok := test.entry.Data[key]
				if !ok {
					t.Errorf("\"%s\" was not expected", key)
				}
				if !expectedValue(value) {
					t.Errorf("wrong value for \"%s\": got %T.%v", key, value, value)
				}
			}
		})
	}
}

func TestWithoutLogID(t *testing.T) {
	tests := []struct {
		name           string
		entry          *Entry
		expectedFields map[string]func(interface{}) bool
	}{
		{
			"without error",
			New(),
			map[string]func(interface{}) bool{
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
			},
		},
		{
			"with error",
			New().WithError(errTest),
			map[string]func(interface{}) bool{
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"error": func(got interface{}) bool {
					err, ok := got.(error)
					if !ok {
						return false
					}
					return errors.Is(err, errTest)
				},
			},
		},
		{
			"on error",
			OnError(errTest),
			map[string]func(interface{}) bool{
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"error": func(got interface{}) bool {
					err, ok := got.(error)
					if !ok {
						return false
					}
					return errors.Is(err, errTest)
				},
			},
		},
		{
			"on error without",
			OnError(nil),
			map[string]func(interface{}) bool{},
		},
		{
			"with fields",
			WithFields("field1", 134, "field2", "asdlkfj"),
			map[string]func(interface{}) bool{
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"field1": func(got interface{}) bool {
					i, ok := got.(int)
					if !ok {
						return false
					}
					return i == 134
				},
				"field2": func(got interface{}) bool {
					i, ok := got.(string)
					if !ok {
						return false
					}
					return i == "asdlkfj"
				},
			},
		},
		{
			"with field",
			WithFields().WithField("field1", 134),
			map[string]func(interface{}) bool{
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"field1": func(got interface{}) bool {
					i, ok := got.(int)
					if !ok {
						return false
					}
					return i == 134
				},
			},
		},
		{
			"fields odd",
			WithFields("kevin"),
			map[string]func(interface{}) bool{
				"caller": func(got interface{}) bool {
					s, ok := got.(string)
					if !ok {
						return false
					}
					return strings.Contains(s, "logging/logging_test.go:")
				},
				"oddFields": func(got interface{}) bool {
					i, ok := got.(int)
					if !ok {
						return false
					}
					return i == 1
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.entry.Debug()
			if len(test.entry.Data) != len(test.expectedFields) {
				t.Errorf("enexpected amount of fields got: %d, want %d", len(test.entry.Data), len(test.expectedFields))
			}
			for key, expectedValue := range test.expectedFields {
				value, ok := test.entry.Data[key]
				if !ok {
					t.Errorf("\"%s\" was not expected", key)
				}
				if !expectedValue(value) {
					t.Errorf("wrong value for \"%s\": got %T.%v", key, value, value)
				}
			}
		})
	}
}
