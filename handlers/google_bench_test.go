package handlers_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/zitadel/logging/handlers"
)

func BenchmarkHandlers(b *testing.B) {
	logTime := time.Now()

	// Create a test record
	record := slog.Record{
		Time:    logTime,
		Level:   slog.LevelInfo,
		Message: "Benchmarking log message",
	}
	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	// Benchmark GoogleHandler
	b.Run("GoogleHandler", func(b *testing.B) {
		var buf bytes.Buffer
		handler := handlers.NewGoogle(&buf, nil, nil).WithAttrs(attrs)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			if err := handler.Handle(context.Background(), record); err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
			buf.Reset() // Clear buffer for the next iteration
		}
	})

	// Benchmark slog.JSONHandler
	b.Run("JSONHandler", func(b *testing.B) {
		var buf bytes.Buffer
		handler := slog.NewJSONHandler(&buf, nil).WithAttrs(attrs)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			if err := handler.Handle(context.Background(), record); err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
			buf.Reset() // Clear buffer for the next iteration
		}
	})
}
