package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Entry struct {
	attributes   []slog.Attr
	isOnError    bool
	err          error
	ts           time.Time
	skipToCaller int
}

var idKey = "logID"

// SetIDKey key of id in logentry
func SetIDKey(key string) {
	idKey = key
}

// Deprecated: Log creates a new entry with an id
func Log(id string) *Entry {
	return &Entry{attributes: []slog.Attr{slog.String(idKey, id)}}
}

// Deprecated: LogWithFields creates a new entry with an id and the given fields
func LogWithFields(id string, fields ...interface{}) *Entry {
	return Log(id).SetFields(fields...)
}

// New instantiates a new entry
func New() *Entry { return &Entry{} }

func OnError(err error) *Entry {
	return New().OnError(err)
}

func WithError(err error) *Entry {
	return New().WithError(err)
}

// WithFields creates a new entry without an id and the given fields
func WithFields(fields ...interface{}) *Entry {
	return New().SetFields(fields...)
}

// OnError sets the error. The log will only be printed if err is not nil
func (e *Entry) OnError(err error) *Entry {
	e.WithError(err)
	e.isOnError = true
	return e
}

// SetFields sets the given fields on the entry. It panics if length of fields is odd
func (e *Entry) SetFields(fields ...interface{}) *Entry {
	logFields := toFields(fields...)
	e.attributes = append(e.attributes, logFields...)
	return e
}

func (e *Entry) WithField(key string, value interface{}) *Entry {
	return e.WithFields(map[string]interface{}{key: value})
}

func (e *Entry) WithFields(fields map[string]any) *Entry {
	for k, v := range fields {
		e.attributes = append(e.attributes, slog.Any(k, v))
	}
	return e
}

func (e *Entry) WithError(err error) *Entry {
	e.err = err
	if err == nil {
		return e
	}
	for i := range e.attributes {
		attr := e.attributes[i]
		if attr.Key == "err" {
			return e
		}
	}
	e.attributes = append(e.attributes, slog.Any("err", err))
	return e
}

func (e *Entry) WithTime(t time.Time) *Entry {
	e.ts = t
	return e
}

func Debug(args ...interface{}) {
	Debugln(args...)
}

func (e *Entry) Debug(args ...interface{}) {
	e.Debugln(args...)
}

func Debugln(args ...interface{}) {
	New().Debugln(args...)
}

func (e *Entry) Debugln(args ...interface{}) {
	msg, attrs := slogArgs(args)
	e.log(func(logger *slog.Logger) {
		logger.Debug(msg, anyArgs(append(e.attributes, attrs...))...)
	})
}

func Debugf(format string, args ...interface{}) {
	New().Debugf(format, args...)
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.log(func(l *slog.Logger) {
		slog.Default().Debug(slogArgsf(format, args...), anyArgs(e.attributes)...)
	})
}

func Info(args ...interface{}) {
	Infoln(args...)
}

func (e *Entry) Info(args ...interface{}) {
	e.Infoln(args...)
}

func Infoln(args ...interface{}) {
	New().Infoln(args...)
}

func (e *Entry) Infoln(args ...interface{}) {
	msg, attrs := slogArgs(args)
	e.log(func(l *slog.Logger) {
		l.Info(msg, anyArgs(append(e.attributes, attrs...))...)
	})
}

func Infof(format string, args ...interface{}) {
	New().Infof(format, args...)
}

func (e *Entry) Infof(format string, args ...interface{}) {
	e.log(func(l *slog.Logger) {
		l.Info(slogArgsf(format, args...), anyArgs(e.attributes)...)
	})
}

func Trace(args ...interface{}) {
	Traceln(args...)
}

func (e *Entry) Trace(args ...interface{}) {
	e.Traceln(args...)
}

func Traceln(args ...interface{}) {
	New().Traceln(args...)
}

func (e *Entry) Traceln(args ...interface{}) {
	msg, attrs := slogArgs(args)
	e.log(func(l *slog.Logger) {
		l.Log(context.Background(), -8, msg, anyArgs(append(e.attributes, attrs...))...)
	})
}

func Tracef(format string, args ...interface{}) {
	New().Tracef(format, args...)
}

func (e *Entry) Tracef(format string, args ...interface{}) {
	e.log(func(l *slog.Logger) {
		l.Log(context.Background(), -8, slogArgsf(format, args...), anyArgs(e.attributes)...)
	})
}

func Warn(args ...interface{}) {
	Warnln(args...)
}

func (e *Entry) Warn(args ...interface{}) {
	e.Warnln(args...)
}

func Warnln(args ...interface{}) {
	New().Warnln(args...)
}

func (e *Entry) Warnln(args ...interface{}) {
	msg, attrs := slogArgs(args)
	e.log(func(l *slog.Logger) {
		l.Warn(msg, anyArgs(append(e.attributes, attrs...))...)
	})
}

func Warnf(format string, args ...interface{}) {
	New().Warnf(format, args...)
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	e.log(func(l *slog.Logger) {
		l.Warn(slogArgsf(format, args...), anyArgs(e.attributes)...)
	})
}

func Warning(args ...interface{}) {
	Warn(args...)
}

func (e *Entry) Warning(args ...interface{}) {
	e.Warn(args...)
}

func Warningln(args ...interface{}) {
	Warnln(args...)
}

func (e *Entry) Warningln(args ...interface{}) {
	e.Warnln(args...)
}

func Warningf(format string, args ...interface{}) {
	Warnf(format, args...)
}

func (e *Entry) Warningf(format string, args ...interface{}) {
	e.Warnf(format, args...)
}

func Error(args ...interface{}) {
	Errorln(args...)
}

func (e *Entry) Error(args ...interface{}) {
	e.Errorln(args...)
}

func Errorln(args ...interface{}) {
	New().Errorln(args...)
}

func (e *Entry) Errorln(args ...interface{}) {
	e.log(func(l *slog.Logger) {
		msg, attrs := slogArgs(args)
		l.Error(msg, anyArgs(append(e.attributes, attrs...))...)
	})
}

func Errorf(format string, args ...interface{}) {
	New().Errorf(format, args...)
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.log(func(l *slog.Logger) {
		slog.Error(slogArgsf(format, args...), anyArgs(e.attributes)...)
	})
}

func Fatal(args ...interface{}) {
	Fatalln(args...)
}

func (e *Entry) Fatal(args ...interface{}) {
	e.Fatalln(args...)
}

func Fatalln(args ...interface{}) {
	New().Fatalln(args...)
}

func (e *Entry) Fatalln(args ...interface{}) {
	msg, attrs := slogArgs(args)
	e.log(func(l *slog.Logger) {
		l.Log(context.Background(), 12, msg, anyArgs(append(e.WithError(errors.New(msg)).attributes, attrs...))...)
		os.Exit(1)
	})
}

func Fatalf(format string, args ...interface{}) {
	New().Fatalf(format, args...)
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	msg := slogArgsf(format, args...)
	e.log(func(l *slog.Logger) {
		l.Log(context.Background(), 12, msg, anyArgs(e.WithError(errors.New(msg)).attributes)...)
		os.Exit(1)
	})
}

func Panic(args ...interface{}) {
	Panicln(args...)
}

func (e *Entry) Panic(args ...interface{}) {
	e.Panicln(args...)
}

func Panicln(args ...interface{}) {
	New().Panicln(args...)
}

func (e *Entry) Panicln(args ...interface{}) {
	msg, attrs := slogArgs(args)
	e.log(func(l *slog.Logger) {
		l.Log(context.Background(), 16, msg, anyArgs(append(e.WithError(errors.New(msg)).attributes, attrs...))...)
		panic(msg)
	})
}

func Panicf(format string, args ...interface{}) {
	Fatalf(format, args...)
}

func (e *Entry) Panicf(format string, args ...interface{}) {
	msg := slogArgsf(format, args...)
	e.log(func(l *slog.Logger) {
		l.Log(context.Background(), 16, msg, anyArgs(e.WithError(errors.New(msg)).attributes)...)
		panic(msg)
	})
}

func (e *Entry) log(log func(logger *slog.Logger)) {
	e = e.checkOnError()
	if e == nil {
		return
	}
	log(slog.Default())
}

func (e *Entry) checkOnError() *Entry {
	if !e.isOnError {
		return e
	}
	if e.err == nil {
		return nil
	}
	return e
}

func slogArgs(args []any) (string, []slog.Attr) {
	if len(args) == 0 {
		return "", nil
	}
	msg, ok := args[0].(string)
	if !ok {
		msg = fmt.Sprintf("%+v", args[0])
		args = append(args, "nonstringloggingkey", fmt.Sprintf("%+v of type %T", args[0], args[0]))
	}
	return msg, toFields(args[1:]...)
}

func toFields(fields ...interface{}) []slog.Attr {
	if len(fields)%2 != 0 {
		return []slog.Attr{slog.Int("oddFields", len(fields))}
	}
	logFields := make([]slog.Attr, 0, len(fields)/2)
	for i := 0; i < len(fields); i = i + 2 {
		key := fields[i].(string)
		logFields = append(logFields, slog.Any(key, fields[i+1]))
	}
	return logFields
}

func anyArgs(attrs []slog.Attr) (anys []any) {
	for _, attr := range attrs {
		anys = append(anys, attr)
	}
	return
}

var slogArgsf = fmt.Sprintf
