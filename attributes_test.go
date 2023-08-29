package logging

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_requestToAttr(t *testing.T) {
	out, logger := newTestLogger()
	logger.Info("test", requestToAttr(
		httptest.NewRequest("GET", "/taget", nil),
	))

	want := `{
		"level":"INFO",
		"msg":"test",
		"time":"not",
		"request":{
			"method":"GET",
			"url":"/taget"
		}
	}`
	got := out.String()
	assert.JSONEq(t, want, got)
}
