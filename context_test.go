package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestContext(t *testing.T) {
	got, ok := FromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, got)

	want := slog.Default()
	ctx := ToContext(context.Background(), want)
	got, ok = FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, want, got)
}
