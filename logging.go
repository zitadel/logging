package logging

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Entry struct {
	*logrus.Entry
	err error
}

var idKey = "logID"

// SetIDKey key of id in logentry
func SetIDKey(key string) {
	idKey = key
}

// Log creates a new entry with an id
func Log(id string) *Entry {
	entry := (*logrus.Logger)(log).WithField(idKey, id)
	entry.Logger = (*logrus.Logger)(log)
	return &Entry{Entry: entry}
}

// OnError sets the error. The log will only be printed if err is not nil
func (e *Entry) OnError(err error) *Entry {
	e.err = err
	return e
}

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

func (e *Entry) Debug(args ...interface{}) {
	e.Log(logrus.DebugLevel, args...)
}

func (e *Entry) Debugln(args ...interface{}) {
	e.Logln(logrus.DebugLevel, args...)
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.Logf(logrus.DebugLevel, format, args...)
}

func (e *Entry) Info(args ...interface{}) {
	e.Log(logrus.InfoLevel, args...)
}

func (e *Entry) Infoln(args ...interface{}) {
	e.Logln(logrus.InfoLevel, args...)
}

func (e *Entry) Infof(format string, args ...interface{}) {
	e.Logf(logrus.InfoLevel, format, args...)
}

func (e *Entry) Trace(args ...interface{}) {
	e.Log(logrus.TraceLevel, args...)
}

func (e *Entry) Traceln(args ...interface{}) {
	e.Logln(logrus.TraceLevel, args...)
}

func (e *Entry) Tracef(format string, args ...interface{}) {
	e.Logf(logrus.TraceLevel, format, args...)
}

func (e *Entry) Warn(args ...interface{}) {
	e.Log(logrus.WarnLevel, args...)
}

func (e *Entry) Warnln(args ...interface{}) {
	e.Logln(logrus.WarnLevel, args...)
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	e.Logf(logrus.WarnLevel, format, args...)
}

func (e *Entry) Warning(args ...interface{}) {
	e.Log(logrus.WarnLevel, args...)
}

func (e *Entry) Warningln(args ...interface{}) {
	e.Logln(logrus.WarnLevel, args...)
}

func (e *Entry) Warningf(format string, args ...interface{}) {
	e.Logf(logrus.WarnLevel, format, args...)
}

func (e *Entry) Error(args ...interface{}) {
	e.Log(logrus.ErrorLevel, args...)
}

func (e *Entry) Errorln(args ...interface{}) {
	e.Logln(logrus.ErrorLevel, args...)
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.Logf(logrus.ErrorLevel, format, args...)
}

func (e *Entry) Fatal(args ...interface{}) {
	e.Log(logrus.FatalLevel, args...)
}

func (e *Entry) Fatalln(args ...interface{}) {
	e.Logln(logrus.FatalLevel, args...)
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.Logf(logrus.FatalLevel, format, args...)
}

func (e *Entry) Panic(args ...interface{}) {
	e.Log(logrus.PanicLevel, args...)
}

func (e *Entry) Panicln(args ...interface{}) {
	e.Logln(logrus.PanicLevel, args...)
}

func (e *Entry) Panicf(format string, args ...interface{}) {
	e.Logf(logrus.PanicLevel, format, args...)
}

func (e *Entry) Log(level logrus.Level, args ...interface{}) {
	e.setError().Entry.Log(level, args...)
}

func (e *Entry) Logf(level logrus.Level, format string, args ...interface{}) {
	e.setError().Entry.Logf(level, format, args...)
}

func (e *Entry) Logln(level logrus.Level, args ...interface{}) {
	e.setError().Entry.Logln(level, args...)
}

func (e *Entry) setError() *Entry {
	if e.err != nil {
		e.WithError(e.err)
	}
	return e
}
