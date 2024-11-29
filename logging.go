package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type Entry struct {
	attributes []any
	isOnError  bool
	err        error
	ts         time.Time
}

var idKey = "logID"

// SetIDKey key of id in logentry
func SetIDKey(key string) {
	idKey = key
}

// Deprecated: Log creates a new entry with an id
func Log(id string) *Entry {
	return &Entry{attributes: []any{slog.String(idKey, id)}}
}

// Deprecated: LogWithFields creates a new entry with an id and the given fields
func LogWithFields(id string, fields ...interface{}) *Entry {
	return Log(id).SetFields(fields...)
}

// New instantiates a new entry
func New() *Entry {
	return &Entry{}
}

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
	e.attributes = logFields
	return e
}

func (e *Entry) WithField(key string, value interface{}) *Entry {
	return e.WithFields(map[string]interface{}{key: value})
}

func (e *Entry) WithFields(fields map[string]interface{}) *Entry {
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
	for _, attr := range e.attributes {
		if a, ok := attr.(slog.Attr); ok && a.Key == "err" {
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

func toFields(fields ...interface{}) []any {
	if len(fields)%2 != 0 {
		return []any{slog.Int("oddFields", len(fields))}
	}
	logFields := make([]any, 0, len(fields)/2)
	for i := 0; i < len(fields); i = i + 2 {
		key := fields[i].(string)
		logFields = append(logFields, slog.Any(key, fields[i+1]))
	}
	return logFields
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
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Debug(msg, append(e.attributes, attrs...)...)
	})
}

func Debugf(format string, args ...interface{}) {
	New().Debugf(format, args...)
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Debug(slogArgsf(format, args...), e.attributes...)
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
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Info(msg, append(e.attributes, attrs...)...)
	})
}

func Infof(format string, args ...interface{}) {
	New().Infof(format, args...)
}

func (e *Entry) Infof(format string, args ...interface{}) {
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Info(slogArgsf(format, args...), e.attributes...)
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
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Log(context.Background(), -8, msg, append(e.attributes, attrs...)...)
	})
}

func Tracef(format string, args ...interface{}) {
	New().Tracef(format, args...)
}

func (e *Entry) Tracef(format string, args ...interface{}) {
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Log(context.Background(), -8, slogArgsf(format, args...), e.attributes...)
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
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Warn(msg, append(e.attributes, attrs...)...)
	})
}

func Warnf(format string, args ...interface{}) {
	New().Warnf(format, args...)
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Warn(slogArgsf(format, args...), e.attributes...)
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

type customErrorFormatter struct {
	next slog.Handler
}

func (c customErrorFormatter) Enabled(context.Context, slog.Level) bool {
	return true
}

func (c customErrorFormatter) Handle(ctx context.Context, record slog.Record) error {
	newAttrs := []slog.Attr{
		slog.String("severity", strings.ToUpper(record.Level.String())),
		slog.String("time", record.Time.Format(time.RFC3339)),
	}
	var errField string
	record.Attrs(func(attr slog.Attr) bool {
		switch attr.Key {
		case "err":
			errField = attr.Value.String()
			newAttrs = append(newAttrs,
				slog.String("@type", "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"),
				slog.String("stack_trace", string(debug.Stack())),
			)
		case "level":
			newAttrs = append(newAttrs, slog.String("severity", strings.ToUpper(attr.Value.String())))
		case "msg":
			// filter
		default:
			newAttrs = append(newAttrs, attr)
		}
		return true
	})
	msg := record.Message
	if errField != "" {
		msg = fmt.Sprintf("%s: %s", record.Message, errField)
	}
	newAttrs = append(newAttrs, slog.String("message", msg))
	newRecord := slog.NewRecord(record.Time, record.Level, "", record.PC)
	newRecord.AddAttrs(newAttrs...)
	return c.next.Handle(ctx, newRecord)
}

func (c customErrorFormatter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return customErrorFormatter{c.next.WithAttrs(attrs)}
}

func (c customErrorFormatter) WithGroup(name string) slog.Handler {
	return customErrorFormatter{c.next.WithGroup(name)}
}

func (e *Entry) Errorln(args ...interface{}) {
	e.log(func() {
		msg, attrs := slogArgs(args)
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Error(msg, append(e.attributes, attrs...)...)
	})
}

func Errorf(format string, args ...interface{}) {
	New().Errorf(format, args...)
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.log(func() { slog.Error(slogArgsf(format, args...), e.attributes...) })
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
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Log(context.Background(), 12, msg, append(e.WithError(errors.New(msg)).attributes, attrs...)...)
		os.Exit(1)
	})
}

func Fatalf(format string, args ...interface{}) {
	New().Fatalf(format, args...)
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	msg := slogArgsf(format, args...)
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Log(context.Background(), 12, msg, e.WithError(errors.New(msg)).attributes...)
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
	e.log(func() {
		e.WithError(errors.New(msg))
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Log(context.Background(), 16, msg, append(e.WithError(errors.New(msg)).attributes, attrs...)...)
		panic(msg)
	})
}

func Panicf(format string, args ...interface{}) {
	Fatalf(format, args...)
}

func (e *Entry) Panicf(format string, args ...interface{}) {
	msg := slogArgsf(format, args...)
	e.log(func() {
		slog.New(customErrorFormatter{next: slog.Default().Handler()}).Log(context.Background(), 16, msg, e.WithError(errors.New(msg)).attributes...)
		panic(msg)
	})
}

func (e *Entry) log(log func()) {
	e = e.checkOnError()
	if e == nil {
		return
	}
	addCaller(e)
	log()
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

func addCaller(e *Entry) {
	_, file, no, ok := runtime.Caller(3)
	if ok {
		e.WithField("caller", fmt.Sprintf("%s:%d", file, no))
	}
}

func slogArgs(args []interface{}) (string, []interface{}) {
	msg, ok := args[0].(string)
	if !ok {
		msg = fmt.Sprintf("%+v", args[0])
		args = append(args, "nonstringloggingkey", fmt.Sprintf("%+v of type %T", args[0]))
	}
	return msg, args[1:]
}

var slogArgsf = fmt.Sprintf
