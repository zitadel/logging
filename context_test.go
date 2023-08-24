package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

// Implementation check
var _ = slog.LogValuer(ContextData{})

func TestNewContextData(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want ContextData
	}{
		{
			name: "without id",
			args: args{},
			want: ContextData{},
		},
		{
			name: "with id",
			args: args{"id1"},
			want: ContextData{
				ContextDataIDKey: "id1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewContextData(tt.args.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContextData_LogValue(t *testing.T) {
	want := slog.GroupValue(
		slog.Any("foo", "bar"),
		slog.Any("hello", "world"),
	)
	data := ContextData{
		"hello": "world",
		"foo":   "bar",
	}
	got := data.LogValue()
	assert.Equal(t, want, got)
}

func TestContextData(t *testing.T) {
	data := ContextData{
		"hello": "world",
	}
	ctx := ContextWithData(context.Background(), data)
	got, ok := DataFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, data, got)

	ctx = ContextWithData(ctx, ContextData{"foo": "bar"})
	want := ContextData{
		"hello": "world",
		"foo":   "bar",
	}
	got, ok = DataFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, want, got)

	ctx = ContextWithData(ctx, ContextData{"hello": "other"})
	want = ContextData{
		"hello": "other",
		"foo":   "bar",
	}
	got, ok = DataFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, want, got)
}
