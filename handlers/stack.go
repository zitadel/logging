package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

const skipKey = "stack_handler_internal_caller_skip"

var _ slog.Handler = (*StackHandler)(nil)

type StackHandler struct {
	wrapped slog.Handler
}

// AddCallerAndStack wraps a wrapped and adds a stack trace to the log entry.
func AddCallerAndStack(handler slog.Handler) *StackHandler {
	return &StackHandler{wrapped: handler}
}

func (s *StackHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return s.wrapped.Enabled(ctx, level)
}

func (s *StackHandler) Handle(ctx context.Context, record slog.Record) error {
	record = withCaller(record)
	record = replaceInternalKey(record, record.Level >= slog.LevelError)
	return s.wrapped.Handle(ctx, record)
}

func (s *StackHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return s.wrapped.WithAttrs(attrs)
}

func (s *StackHandler) WithGroup(name string) slog.Handler {
	return s.wrapped.WithGroup(name)
}

// withCaller adds the library caller to the record.
func withCaller(record slog.Record) slog.Record {
	newRecord := record.Clone()
	// At least skip withCaller and Handle
	skip := 2
	for {
		_, f, l, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		if !skipCaller(f) {
			newRecord.AddAttrs(
				slog.String("caller", fmt.Sprintf("%s:%d", f, l)),
				// Also skip the caller of withCaller
				slog.Int(skipKey, skip+1),
			)
			return newRecord
		}
		skip++
	}
	return record
}

// skipCaller checks if the file belongs to the logging package. However, it doesn't skip test files.
func skipCaller(file string) bool {
	return !strings.Contains(file, "_test.go") && (strings.Contains(file, "github.com/zitadel/logging") ||
		strings.Contains(file, "slog/logger.go"))
}

// if addStack is true, replaceInternalKey generates a stack trace in the same format as debug.Stack(), but skipped until the library caller.
func replaceInternalKey(record slog.Record, addStack bool) slog.Record {
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	var skipToCaller int
	record.Clone().Attrs(func(attr slog.Attr) bool {
		if attr.Key == skipKey {
			skipToCaller = int(attr.Value.Int64())
			return true
		}
		newRecord.AddAttrs(attr)
		return true
	})
	if !addStack {
		return newRecord
	}
	if skipToCaller == 0 {
		newRecord.AddAttrs(slog.String("stack_trace", "stack trace not available, skipToCaller is 0, which means withCaller was not called"))
		return newRecord
	}
	callers := make([]uintptr, 32) // Collect up to 32 stack frames
	n := runtime.Callers(skipToCaller, callers)
	frames := runtime.CallersFrames(callers[:n])
	var buf bytes.Buffer
	for {
		frame, more := frames.Next()
		buf.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	newRecord.AddAttrs(slog.String("stack_trace", buf.String()))
	return newRecord
}
