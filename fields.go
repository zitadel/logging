package logging

import (
	"time"

	"github.com/sirupsen/logrus"
)

// LogWithFields creates a new entry with an id and the given fields
func LogWithFields(id string, fields ...interface{}) *Entry {
	e := Log(id)
	return e.SetFields(fields...)
}

// SetFields sets the given fields on the entry. It panics if length of fields is odd
func (e *Entry) SetFields(fields ...interface{}) *Entry {
	logFields := toFields(fields...)
	return e.WithFields(logFields)
}

func (e *Entry) WithField(key string, value interface{}) *Entry {
	e.Entry = e.Entry.WithField(key, value)
	return e
}

func (e *Entry) WithFields(fields logrus.Fields) *Entry {
	e.Entry = e.Entry.WithFields(fields)
	return e
}

func (e *Entry) WithError(err error) *Entry {
	e.Entry = e.Entry.WithError(err)
	return e
}

func (e *Entry) WithTime(t time.Time) *Entry {
	e.Entry = e.Entry.WithTime(t)
	return e
}

func toFields(fields ...interface{}) logrus.Fields {
	if len(fields)%2 != 0 {
		return logrus.Fields{"oddFields": len(fields)}
	}
	logFields := make(logrus.Fields, len(fields)%2)
	for i := 0; i < len(fields); i = i + 2 {
		key := fields[i].(string)
		logFields[key] = fields[i+1]
	}
	return logFields
}
