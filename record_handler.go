package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ slog.Handler = (*ZitadelHandler)(nil)

type ZitadelHandler struct {
	w                         io.Writer
	opt                       *slog.HandlerOptions
	service, version, process string
	dynamic                   map[string]any
}

// NewZitadelHandler creates a new ZitadelHandler.
// It writes structured JSON records to w.
func NewZitadelHandler(w io.Writer, opt *slog.HandlerOptions, service, version, process string, dynamic map[string]any) *ZitadelHandler {
	return &ZitadelHandler{
		w:       w,
		opt:     opt,
		service: service,
		version: version,
		process: process,
		dynamic: dynamic,
	}
}

func (z *ZitadelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= z.opt.Level.Level()
}

func (z *ZitadelHandler) WithAttrs([]slog.Attr) slog.Handler {
	return z
}

func (z *ZitadelHandler) WithGroup(string) slog.Handler {
	return z
}

// Handle processes each slog.Record and writes it as a protobuf-serialized message.
func (z *ZitadelHandler) Handle(ctx context.Context, r slog.Record) error {
	protoRecord, err := z.mapRecordToProto(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to map record to proto message: %v", err)
	}
	recordJson, err := protojson.Marshal(protoRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal proto message to JSON: %w", err)
	}
	_, err = z.w.Write(append(recordJson, '\n'))
	return fmt.Errorf("failed to write JSON with trailing newline: %w", err)
}

// toSeverity maps slog.Level to Severity
// Severity implements slog.StringerValuer
func toSeverity(level slog.Level) RecordV1_Severity {
	switch level {
	case slog.LevelDebug:
		return RecordV1_Debug
	case slog.LevelInfo:
		return RecordV1_Info
	case slog.LevelWarn:
		return RecordV1_Warn
	case slog.LevelError:
		return RecordV1_Error
	default:
		return RecordV1_SeverityUndefined
	}
}

const addStackTracePC int = 3

func addStackTrace(slogPC uintptr) []string {
	callers := make([]uintptr, 32) // Collect up to 32 stack frames
	n := runtime.Callers(int(slogPC)+addStackTracePC, callers)
	frames := runtime.CallersFrames(callers[:n])
	stack := make([]string, 0)
	for {
		frame, more := frames.Next()
		stack = append(stack, fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return stack
}

// mapRecordToProto maps slog.Record to a Protobuf Record.
func (z *ZitadelHandler) mapRecordToProto(ctx context.Context, r slog.Record) (*Record, error) {
	severity := toSeverity(r.Level)
	record := &RecordV1{
		Time:     timestamppb.New(r.Time),
		Severity: severity,
		Message:  r.Message,
		Service:  &z.service,
		Version:  &z.version,
		Process:  &z.process,
	}
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()
	if spanCtx.HasTraceID() {
		traceID := spanCtx.TraceID().String()
		record.TraceId = &traceID
	}
	if spanCtx.HasSpanID() {
		spanID := spanCtx.SpanID().String()
		record.SpanId = &spanID
	}
	if z.dynamic != nil {
		for k, v := range z.dynamic {
			record.addDynamic(slog.Any(k, v))
		}
	}
	if severity >= RecordV1_Error || severity == RecordV1_Trace {
		record.StackTrace = addStackTrace(r.PC)
	}
	r.Attrs(func(attr slog.Attr) bool {
		if streamValue, ok := attr.Value.Any().(isRecordV1_Stream); ok {
			record.Stream = streamValue
			return true
		}
		record.addDynamic(attr)
		return true
	})
	return &Record{
		Record: &Record_RecordV1{RecordV1: record},
	}, nil
}

func (r *RecordV1) addDynamic(attr slog.Attr) {
	if r.Dynamic == nil {
		r.Dynamic = &structpb.Struct{}
	}
	if r.Dynamic.Fields == nil {
		r.Dynamic.Fields = make(map[string]*structpb.Value)
	}
	v, err := structpb.NewValue(attr.Value.Any())
	if err != nil {
		v = structpb.NewStringValue(fmt.Sprintf("failed to create a structpb.Value from %T %+v", attr.Value, attr.Value))
	}
	r.Dynamic.Fields[attr.Key] = v
}
