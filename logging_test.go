package logging

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestEntryFields(t *testing.T) {
	tests := []struct {
		name           string
		entry          *Entry
		expectedFields logrus.Fields
	}{
		{
			"without error",
			Log("UTILS-B7l7"),
			logrus.Fields{
				"logID":  "UTILS-B7l7",
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:82",
			},
		},
		{
			"with error",
			Log("UTILS-Ld9V").WithError(fmt.Errorf("im an error")),
			logrus.Fields{
				"logID":  "UTILS-Ld9V",
				"error":  fmt.Errorf("im an error"),
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:82",
			},
		},
		{
			"on error",
			Log("UTILS-Ld9V").OnError(fmt.Errorf("im an error")),
			logrus.Fields{
				"logID":  "UTILS-Ld9V",
				"error":  fmt.Errorf("im an error"),
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:82",
			},
		},
		{
			"on error without",
			Log("UTILS-Ld9V").OnError(nil),
			logrus.Fields{
				"logID":  "UTILS-Ld9V",
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:82",
			},
		},
		{
			"with fields",
			LogWithFields("LOGGI-5kk6z", "field1", 134, "field2", "asdlkfj"),
			logrus.Fields{
				"logID":  "LOGGI-5kk6z",
				"field1": 134,
				"field2": "asdlkfj",
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:82",
			},
		},
		{
			"with field",
			LogWithFields("LOGGI-5kk6z").WithField("field1", 134),
			logrus.Fields{
				"logID":  "LOGGI-5kk6z",
				"field1": 134,
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:82",
			},
		},
		{
			"fields odd",
			LogWithFields("LOGGI-xWzy4", "kevin"),
			logrus.Fields{
				"logID":     "LOGGI-xWzy4",
				"oddFields": 1,
				"caller":    "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:82",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.entry.Debug()
			for key, expectedValue := range test.entry.Data {
				value, ok := test.expectedFields[key]
				if !ok {
					t.Errorf("\"%s\" was not expected", key)
				}
				if !reflect.DeepEqual(expectedValue, value) {
					t.Errorf("wrong value for \"%s\": expected %T.%v, got %T.%v", key, expectedValue, expectedValue, value, value)
				}
			}
		})
	}
}

func TestNewEntryFields(t *testing.T) {
	tests := []struct {
		name           string
		entry          *Entry
		expectedFields logrus.Fields
	}{
		{
			"without error",
			New("UTILS-B7l7"),
			logrus.Fields{
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:160",
			},
		},
		{
			"with error",
			New("UTILS-Ld9V").WithError(fmt.Errorf("im an error")),
			logrus.Fields{
				"error":  fmt.Errorf("im an error"),
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:160",
			},
		},
		{
			"on error",
			New("UTILS-Ld9V").OnError(fmt.Errorf("im an error")),
			logrus.Fields{
				"error":  fmt.Errorf("im an error"),
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:160",
			},
		},
		{
			"on error without",
			New("UTILS-Ld9V").OnError(nil),
			logrus.Fields{
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:160",
			},
		},
		{
			"with fields",
			WithFields("field1", 134, "field2", "asdlkfj"),
			logrus.Fields{
				"field1": 134,
				"field2": "asdlkfj",
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:160",
			},
		},
		{
			"with field",
			WithFields().WithField("field1", 134),
			logrus.Fields{
				"field1": 134,
				"caller": "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:160",
			},
		},
		{
			"fields odd",
			WithFields("kevin"),
			logrus.Fields{
				"oddFields": 1,
				"caller":    "/Users/adlerhurst/go/src/github.com/caos/logging/logging_test.go:160",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.entry.Debug()
			for key, expectedValue := range test.entry.Data {
				value, ok := test.expectedFields[key]
				if !ok {
					t.Errorf("\"%s\" was not expected", key)
				}
				if !reflect.DeepEqual(expectedValue, value) {
					t.Errorf("wrong value for \"%s\": expected %T.%v, got %T.%v", key, expectedValue, expectedValue, value, value)
				}
			}
		})
	}
}
