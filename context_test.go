package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestContext(t *testing.T) {
	want := slog.Default()
	got := FromContext(context.Background())
	assert.Equal(t, want, got)

	ctx := ToContext(context.Background(), want)
	got = FromContext(ctx)
	assert.Equal(t, want, got)
}
