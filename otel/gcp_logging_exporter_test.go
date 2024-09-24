package otel

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
)

func TestNewGCPLoggingExporterHook_InvalidConfig(t *testing.T) {
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/invalid.json"))
	testCases := []struct {
		name      string
		config    func(cfg *googlecloudexporter.Config)
		expectErr bool
	}{
		{
			name: "Invalid Queue Size",
			config: func(cfg *googlecloudexporter.Config) {
				cfg.QueueSize = -1 // invalid queue size
			},
			expectErr: true,
		},
		{
			name: "Invalid Compression",
			config: func(cfg *googlecloudexporter.Config) {
				cfg.LogConfig.ClientConfig.Compression = "invalid" // invalid compression
			},
			expectErr: true,
		},
		{
			name: "Default is valid",
			config: func(cfg *googlecloudexporter.Config) {
				// default config
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hook, err := NewGCPLoggingExporterHook("test-project", WithExporterConfig(tc.config))
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if hook != nil {
				assert.NoError(t, hook.Start())
			}
		})
	}
}

func TestGcpLoggingExporterHook_Fire_DifferentLevels(t *testing.T) {
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/invalid.json"))
	hook, err := NewGCPLoggingExporterHook("test-project")
	require.NoError(t, err)
	require.NoError(t, hook.Start())
	for _, level := range []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel} {
		entry := &logrus.Entry{
			Message: "Test message",
			Level:   level,
		}

		err = hook.Fire(entry)
		assert.NoError(t, err)
	}
}

func TestGcpLoggingExporterHook_Fire_NotStarted(t *testing.T) {
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/invalid.json"))
	hook, err := NewGCPLoggingExporterHook("test-project")
	require.NoError(t, err)
	entry := &logrus.Entry{
		Message: "Test message",
		Level:   logrus.InfoLevel,
	}
	assert.Panics(t, func() { _ = hook.Fire(entry) }, "The code did not panic")
}

func TestGcpLoggingExporterHook_Fire_IncludeExclude(t *testing.T) {
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/invalid.json"))
	hook, err := NewGCPLoggingExporterHook(
		"test-project",
		WithInclude(func(entry *logrus.Entry) bool { return entry.Level == logrus.InfoLevel }),
		WithExclude(func(entry *logrus.Entry) bool { return entry.Message == "Exclude this message" }),
	)
	require.NoError(t, err)
	require.NoError(t, hook.Start())
	entry := &logrus.Entry{
		Message: "Test message",
		Level:   logrus.InfoLevel,
	}
	assert.NoError(t, hook.Fire(entry))
	entry = &logrus.Entry{
		Message: "Exclude this message",
		Level:   logrus.InfoLevel,
	}
	assert.NoError(t, hook.Fire(entry)) // the entry is not sent but it's not an error
}

type MockLogs struct {
	ConsumeLogsFunc func(ctx context.Context, ld plog.Logs) error
}

func (m *MockLogs) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	return m.ConsumeLogsFunc(ctx, ld)
}

func (m *MockLogs) Shutdown(context.Context) error {
	panic("not implemented")
}

func TestExporterWrapper_Export(t *testing.T) {
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/invalid.json"))
	createLogRecord := func(body string, attributes map[string]string, severity log.Severity) sdklog.Record {
		record := sdklog.Record{}
		record.SetBody(log.StringValue(body))
		record.SetSeverity(severity)
		var recordAttributes []log.KeyValue
		for key, value := range attributes {
			recordAttributes = append(recordAttributes, log.String(key, value))
		}
		record.SetAttributes(recordAttributes...)
		return record
	}

	testCases := []struct {
		name           string
		logs           []sdklog.Record
		consumeLogsErr error
		expectErr      bool
	}{
		{
			name:           "No Records",
			logs:           nil,
			consumeLogsErr: nil,
			expectErr:      false,
		},
		{
			name: "Valid Records",
			logs: []sdklog.Record{
				createLogRecord("Test message", map[string]string{"key1": "value1", "key2": "value2"}, log.SeverityInfo),
			},
			consumeLogsErr: nil,
			expectErr:      false,
		},
		{
			name: "ConsumeLogs Returns Error",
			logs: []sdklog.Record{
				createLogRecord("Another test message", map[string]string{"key3": "value3", "key4": "value4"}, log.SeverityDebug),
			},
			consumeLogsErr: errors.New("consume logs error"),
			expectErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogs := &MockLogs{
				ConsumeLogsFunc: func(ctx context.Context, ld plog.Logs) error {
					// Expect that the logs are correctly mapped
					assert.Equal(t, len(tc.logs), ld.ResourceLogs().Len())
					for i, record := range tc.logs {
						logRecord := ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(i)
						assert.Equal(t, record.Body().String(), logRecord.Body().Str())
						assert.Equal(t, record.Severity(), log.Severity(logRecord.SeverityNumber()))
					}
					return tc.consumeLogsErr
				},
			}
			// Replace the real Logs with the mock
			wrapper := &exporterWrapper{
				Logs:      mockLogs,
				zapLogger: zap.NewNop(), // Use a no-op logger for testing
			}
			err := wrapper.Export(context.Background(), tc.logs)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
