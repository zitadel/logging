package logging

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestWrapHandler(t *testing.T) {
	parent := slog.Default().With("a", "b").Handler()
	type args struct {
		handler slog.Handler
		opts    []HandlerOption
	}
	tests := []struct {
		name string
		args args
		want slog.Handler
	}{
		{
			name: "already slogHandler",
			args: args{
				handler: &slogHandler{
					handler: parent,
				},
			},
			want: &slogHandler{
				handler: parent,
			},
		},
		{
			name: "parent handler",
			args: args{
				handler: parent,
			},
			want: &slogHandler{
				handler: parent,
			},
		},
		{
			name: "with CTXGroupName",
			args: args{
				handler: parent,
				opts: []HandlerOption{
					HandlerWithCTXGroupName("ctx"),
				},
			},
			want: &slogHandler{
				handler:      parent,
				ctxGroupName: "ctx",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapHandler(tt.args.handler, tt.args.opts...)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test all the handler methods with log output
func Test_slogHandler_Log(t *testing.T) {
	logOut := new(strings.Builder)
	logger := slog.New(WrapHandler(
		slog.NewJSONHandler(logOut, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
		HandlerWithCTXGroupName("ctx"),
	))

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
