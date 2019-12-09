package logging

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func initBuf() *bytes.Buffer {
	var buf bytes.Buffer
	SetOutput(&buf)
	return &buf
}

func TestLogOutput(t *testing.T) {
	tests := []struct {
		name          string
		log           func(...interface{})
		message       []interface{}
		shouldContain []string
	}{
		{
			"warn without error",
			Log("UTILS-B7l7").Warn,
			[]interface{}{"check", "check"},
			[]string{"UTILS-B7l7", "level=warning", "msg=checkcheck", "logID=UTILS-B7l7"},
		},
		{
			"warn with error",
			Log("UTILS-Ld9V").OnError(fmt.Errorf("im an error")).Warn,
			[]interface{}{"error ocured"},
			[]string{"UTILS-Ld9V", "level=warning", "msg=\"error ocured\"", "error=\"im an error\""},
		},
		{
			"warn with fields",
			LogWithFields("LOGGI-5kk6z", "field1", 134, "field2", "asdlkfj").Warn,
			[]interface{}{"2 fields"},
			[]string{"field1=134", "field2=asdlkfj", "msg=\"2 fields\""},
		},
		{
			"warn with field",
			LogWithFields("LOGGI-5kk6z").WithField("field1", 134).Warn,
			[]interface{}{"1 field"},
			[]string{"field1=134", "msg=\"1 field\""},
		},
		{
			"fields odd",
			LogWithFields("LOGGI-xWzy4", "kevin").Warn,
			[]interface{}{"2 logs expected"},
			[]string{"oddFields=1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := initBuf()
			test.log(test.message...)
			for _, substring := range test.shouldContain {
				if !strings.Contains(buf.String(), substring) {
					t.Errorf("log (%s) must contain %s", buf, substring)
				}
			}
		})
	}
}
