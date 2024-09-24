package otel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/collector/component"
	otelexporter "go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	otellog "go.opentelemetry.io/otel/log"
	noopmeter "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	nooptracer "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

type FilterFunc func(entry *logrus.Entry) bool

type MapBodyFunc func(entry *logrus.Entry) string

type GcpLoggingExporterHook struct {
	logger           otellog.Logger
	zapLogger        *zap.Logger
	levels           []logrus.Level
	factory          otelexporter.Factory
	exporterCfg      *googlecloudexporter.Config
	otelSettings     *otelexporter.Settings
	include, exclude FilterFunc
	add              []otellog.KeyValue
	mapBody          MapBodyFunc
}

type Option func(*GcpLoggingExporterHook)

func WithExporterConfig(changeDefaults func(*googlecloudexporter.Config)) Option {
	return func(g *GcpLoggingExporterHook) {
		changeDefaults(g.exporterCfg)
	}
}

// WithOtelSettings allows to change the default otel settings
func WithOtelSettings(changeDefaults func(*otelexporter.Settings)) Option {
	return func(g *GcpLoggingExporterHook) {
		changeDefaults(g.otelSettings)
	}
}

// WithLevels allows to change the default logrus levels that are exported
// By default, all levels are exported
func WithLevels(levels []logrus.Level) Option {
	return func(g *GcpLoggingExporterHook) {
		g.levels = levels
	}
}

// WithInclude makes sure that only entries, that meet the condition are exported.
// Entries that meet both the WithInclude condition and the WithExclude condition are discarded.
// By default, MatchAllLogs is used
func WithInclude(filter FilterFunc) Option {
	return func(hook *GcpLoggingExporterHook) {
		hook.include = filter
	}
}

// WithExclude makes sure that only entries, that do not meet the condition are exported.
// Entries that meet both the WithInclude condition and the WithExclude condition are discarded.
// By default, MatchAllLogs is used
func WithExclude(filter FilterFunc) Option {
	return func(hook *GcpLoggingExporterHook) {
		hook.exclude = filter
	}
}

// WithAttributes adds attributes to every log entry
func WithAttributes(attributes []otellog.KeyValue) Option {
	return func(hook *GcpLoggingExporterHook) {
		hook.add = attributes
	}
}

// WithMapBody allows to change the default mapping function
// By default, MapMessageToBody is used
func WithMapBody(mapBody MapBodyFunc) Option {
	return func(hook *GcpLoggingExporterHook) {
		hook.mapBody = mapBody
	}
}

var _ FilterFunc = MatchAllLogs

func MatchAllLogs(*logrus.Entry) bool { return true }

var _ FilterFunc = MatchAllLogs

func MatchNoLogs(*logrus.Entry) bool { return false }

var _ MapBodyFunc = MapMessageToBody

// MapMessageToBody maps logrus entry message to otel log body
func MapMessageToBody(entry *logrus.Entry) string {
	return entry.Message
}

// MapFieldsToBody maps logrus entry fields to otel log body
// This is the default mapping function
func MapFieldsToBody(entry *logrus.Entry) string {
	entryCopy := *entry
	lg := logrus.New()
	lg.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableQuote:     true,
		DisableTimestamp: true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
	}
	entryCopy.Logger = lg
	body, err := entryCopy.String()
	if err != nil {
		return fmt.Sprintf("Failed to format log entry: %v, original message: %s", err, entry.Message)
	}
	return strings.TrimSpace(body)
}

/*
NewGCPLoggingExporterHook creates a new GCPLoggingExporterHook. This hook can be used for exporting logs created with logrus to Google Cloud Logging. The logs are also still printed to stdout.
- projectID is the GCP project ID
- options are used to change the default settings

Make sure the runtime has Google Application Default Credentials set up.

The hook is not started automatically, Start() must be called to start the hook.

Example usage:

```go
logger := logrus.New()
hook, err := NewGCPLoggingExporterHook(

	"your-gcp-project-id",
	WithExporterConfig(func(cfg *googlecloudexporter.Config) {
		cfg.LogConfig.DefaultLogName = "zitadel"
		cfg.LogConfig.ServiceResourceLabels = false
	    cfg.QueueSize = 20
	}),
	WithOtelSettings(func(settings *otelexporter.Settings) {
	    settings.ID = component.MustNewID("custom-id")
	}),
	WithLevels([]logrus.Level{logrus.InfoLevel, logrus.ErrorLevel}),
	WithInclude(func(entry *logrus.Entry) bool {
	    return entry.Data["stream"] == "activity"
	}),
	WithExclude(func(entry *logrus.Entry) bool {
	    return entry.Message == "Exclude this message"
	}),
	WithAttributes([]otellog.KeyValue{
	    otellog.String("key1", "value1"),
	    otellog.String("key2", "value2"),
	}),
	WithMapBody(MapMessageToBody),

)

	if err != nil {
		log.Fatalf("Failed to create GCPLoggingExporterHook: %v", err)
	}

err = hook.Start()

	if err != nil {
		log.Fatalf("Failed to start GCPLoggingExporterHook: %v", err)
	}

logger.AddHook(hook)
```
*/
func NewGCPLoggingExporterHook(projectID string, options ...Option) (*GcpLoggingExporterHook, error) {
	factory := googlecloudexporter.NewFactory()
	cfg := factory.CreateDefaultConfig()
	exporterCfg := cfg.(*googlecloudexporter.Config)
	exporterCfg.LogConfig.DefaultLogName = "default"
	exporterCfg.QueueSize = 10
	exporterCfg.NumConsumers = 1
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	settings := &otelexporter.Settings{
		ID: component.MustNewID("gcp_logging_exporter_logrus_hook"),
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zapLogger,
			MeterProvider:  noopmeter.NewMeterProvider(),
			TracerProvider: nooptracer.NewTracerProvider(),
		},
	}
	hook := &GcpLoggingExporterHook{
		factory:      factory,
		exporterCfg:  exporterCfg,
		otelSettings: settings,
		levels:       logrus.AllLevels,
		zapLogger:    zapLogger,
		include:      MatchAllLogs,
		exclude:      MatchNoLogs,
		mapBody:      MapFieldsToBody,
	}
	for _, option := range options {
		option(hook)
	}
	hook.exporterCfg.ProjectID = projectID
	if err = hook.exporterCfg.Validate(); err != nil {
		return nil, err
	}
	if err = hook.exporterCfg.QueueSettings.Validate(); err != nil {
		return nil, err
	}
	return hook, nil
}

func (o *GcpLoggingExporterHook) Start() error {
	exporter, err := o.factory.CreateLogsExporter(context.Background(), *o.otelSettings, o.exporterCfg)
	if err != nil {
		return err
	}
	if err = exporter.Start(context.Background(), nil); err != nil {
		return err
	}
	logProvider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(&exporterWrapper{zapLogger: o.zapLogger, Logs: exporter})))
	o.logger = logProvider.Logger("hook")
	return nil
}

func (o *GcpLoggingExporterHook) Levels() []logrus.Level {
	return o.levels
}

func (o *GcpLoggingExporterHook) Fire(entry *logrus.Entry) error {
	if o.logger == nil {
		panic("hook not started")
	}
	if !o.include(entry) || o.exclude(entry) {
		return nil
	}
	r := &otellog.Record{}
	r.SetBody(otellog.StringValue(o.mapBody(entry)))
	r.SetTimestamp(time.Now())
	r.SetSeverity(mapLogrusLevelToSeverity(entry.Level))
	r.SetObservedTimestamp(entry.Time)
	r.SetSeverityText(entry.Level.String())
	attrs := make([]otellog.KeyValue, 0)
	for key, value := range entry.Data {
		if value == nil {
			continue
		}
		attrs = append(attrs, otellog.KeyValue{
			Key:   key,
			Value: mapValueToAttributeValue(value),
		})
	}
	attrs = append(attrs, o.add...)
	r.AddAttributes(attrs...)
	o.logger.Emit(context.Background(), *r)
	return nil
}

func mapValueToAttributeValue(value any) otellog.Value {
	switch v := value.(type) {
	case string:
		return otellog.StringValue(v)
	case int:
		return otellog.IntValue(v)
	case int64:
		return otellog.IntValue(int(v))
	case float64:
		return otellog.Float64Value(v)
	case bool:
		return otellog.BoolValue(v)
	default:
		return otellog.StringValue(fmt.Sprintf("%v", v))
	}
}

func mapLogrusLevelToSeverity(level logrus.Level) otellog.Severity {
	switch level {
	case logrus.TraceLevel:
		return otellog.SeverityTrace
	case logrus.DebugLevel:
		return otellog.SeverityDebug
	case logrus.InfoLevel:
		return otellog.SeverityInfo
	case logrus.WarnLevel:
		return otellog.SeverityWarn
	case logrus.ErrorLevel:
		return otellog.SeverityError
	case logrus.FatalLevel, logrus.PanicLevel:
		return otellog.SeverityFatal
	default:
		return otellog.SeverityUndefined
	}
}

var _ sdklog.Exporter = (*exporterWrapper)(nil)

type Logs interface {
	ConsumeLogs(ctx context.Context, ld plog.Logs) error
	Shutdown(ctx context.Context) error
}

type exporterWrapper struct {
	Logs
	zapLogger *zap.Logger
}

func (e *exporterWrapper) Export(ctx context.Context, records []sdklog.Record) error {
	if len(records) == 0 {
		return nil
	}
	ld := plog.NewLogs()
	logResourceLogs := ld.ResourceLogs().AppendEmpty()
	scopeLogs := logResourceLogs.ScopeLogs().AppendEmpty()
	for _, record := range records {
		logRecord := scopeLogs.LogRecords().AppendEmpty()
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(record.Timestamp()))
		logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(record.ObservedTimestamp()))
		logRecord.Body().SetStr(record.Body().String())
		logRecord.SetSeverityText(record.SeverityText())
		logRecord.SetSeverityNumber(plog.SeverityNumber(record.Severity()))
		logRecordAttributes := logRecord.Attributes()
		record.WalkAttributes(func(kv otellog.KeyValue) bool {
			logRecordAttributes.PutStr(kv.Key, kv.Value.String())
			return true
		})
	}
	return e.ConsumeLogs(ctx, ld)
}

func (e *exporterWrapper) ForceFlush(context.Context) error {
	return fmt.Errorf("not implemented")
}
