package logging

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestWrapLogger(t *testing.T) {
	parent := slog.Default().With("a", "b")
	type args struct {
		logger       *slog.Logger
		ctxDataGroup string
	}
	tests := []struct {
		name string
		args args
		want slog.Handler
	}{
		{
			name: "nil logger",
			args: args{nil, ""},
			want: &slogHandler{
				handler: slog.Default().Handler(),
			},
		},
		{
			name: "parent logger",
			args: args{parent, ""},
			want: &slogHandler{
				handler: parent.Handler(),
			},
		},
		{
			name: "ctx data group",
			args: args{parent, "ctx"},
			want: &slogHandler{
				handler:      parent.Handler(),
				ctxDataGroup: "ctx",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapLogger(tt.args.logger, tt.args.ctxDataGroup)
			assert.Equal(t, tt.want, got.Handler())
		})
	}
}

// Test all the handler methods with log output
func Test_slogHandler_Log(t *testing.T) {
	logOut := new(strings.Builder)
	logger := NewLogger(slog.NewJSONHandler(logOut, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}), "ctx")

	// set some data to the context
	ctx := ContextWithData(context.Background(), ContextData{
		"foo":   "bar",
		"hello": "world",
	})

	// We have a InfoLever configured, so LevelDebug should return false
	assert.False(t, logger.Enabled(ctx, slog.LevelDebug))

	logger = logger.With("time", "not") // overwrite the time attribute
	logger = logger.WithGroup("project")
	logger = logger.With("someKey", "someValue")

	logger.InfoContext(ctx, "lets log!")
	want := `{"level":"INFO", "msg":"lets log!", "time":"not", "project":{"someKey":"someValue", "ctx":{"foo":"bar", "hello":"world"}}}`
	got := logOut.String()
	assert.JSONEq(t, want, got)
}
