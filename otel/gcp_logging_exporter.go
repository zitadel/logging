package otel

import (
	"context"
	"fmt"
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

type ContextAttributesFunc func(ctx context.Context) Attributes

type GcpLoggingExporterHook struct {
	logger            otellog.Logger
	levels            []logrus.Level
	factory           otelexporter.Factory
	exporterCfg       *googlecloudexporter.Config
	otelSettings      *otelexporter.Settings
	include, exclude  FilterFunc
	contextAttributes ContextAttributesFunc
	zapLogger         *zap.Logger
}

type Option func(*GcpLoggingExporterHook)

func WithChangedDefaultExporterConfig(changeDefaults func(*googlecloudexporter.Config)) Option {
	return func(g *GcpLoggingExporterHook) {
		changeDefaults(g.exporterCfg)
	}
}

func WithChangedDefaultOtelSettings(changeDefaults func(*otelexporter.Settings)) Option {
	return func(g *GcpLoggingExporterHook) {
		changeDefaults(g.otelSettings)
	}
}

func WithChangedLevels(levels []logrus.Level) Option {
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

// WithContextAttributesFunc is used to extract attributes from an entries context.
// The passed context is nil if WithContext was never called on the log entry
func WithContextAttributesFunc(f ContextAttributesFunc) Option {
	return func(hook *GcpLoggingExporterHook) {
		hook.contextAttributes = f
	}
}

var _ FilterFunc = MatchAllLogs

func MatchAllLogs(*logrus.Entry) bool { return true }

var _ FilterFunc = MatchAllLogs

func MatchNoLogs(*logrus.Entry) bool { return false }

// MatchLogsWithContextKeyFunc returns a FilterFunc that only matches entries
// that have a context with a non-nil value at a certain key
func MatchLogsWithContextKey(key any) FilterFunc {
	return func(entry *logrus.Entry) bool {
		return entry.Context != nil && entry.Context.Value(key) != nil
	}
}

type Attributes []otellog.KeyValue

type Record interface {
	Attributes() Attributes
}

// ContextAttributesFromKey returns a ContextAttributesFunc that extracts the value
// of record interface. If the value is of type Attributes, the attributes are appended to the OTEL log record.
func ContextAttributesFromKeyFunc(key any) ContextAttributesFunc {
	return func(ctx context.Context) Attributes {
		if ctx == nil {
			return nil
		}
		if a, ok := ctx.Value(key).(Record); ok {
			return a.Attributes()
		}
		return nil
	}
}

// NoContextAttributes is a ContextAttributesFunc that doesn't extract any attributes from the context.
// This is the default behavior.
func NoContextAttributes(context.Context) Attributes { return nil }

func NewGCPLoggingExporterHook(options ...Option) (*GcpLoggingExporterHook, error) {
	factory := googlecloudexporter.NewFactory()
	cfg := factory.CreateDefaultConfig()
	exporterCfg := cfg.(*googlecloudexporter.Config)
	exporterCfg.LogConfig.DefaultLogName = "default"
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
		factory:           factory,
		exporterCfg:       exporterCfg,
		otelSettings:      settings,
		levels:            logrus.AllLevels,
		zapLogger:         zapLogger,
		include:           MatchAllLogs,
		exclude:           MatchNoLogs,
		contextAttributes: NoContextAttributes,
	}
	for _, option := range options {
		option(hook)
	}
	if hook.exporterCfg.Validate() != nil {
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
	logProvider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(&exporterWrapper{exporter, o.zapLogger})))
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
	r.SetBody(otellog.StringValue(entry.Message))
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
	r.AddAttributes(append(attrs, o.contextAttributes(entry.Context)...)...)
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

type exporterWrapper struct {
	otelexporter.Logs
	zapLogger *zap.Logger
}

func (e *exporterWrapper) Export(ctx context.Context, records []sdklog.Record) error {
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
			return mapAttributes(kv, logRecordAttributes)
		})
	}
	return e.ConsumeLogs(ctx, ld)
}

func (e *exporterWrapper) ForceFlush(context.Context) error {
	return fmt.Errorf("not implemented")
}

func mapAttributes(kv otellog.KeyValue, into pcommon.Map) bool {
	switch kv.Value.Kind() {
	case otellog.KindString:
		into.PutStr(kv.Key, kv.Value.AsString())
	case otellog.KindInt64:
		into.PutInt(kv.Key, kv.Value.AsInt64())
	case otellog.KindBool:
		into.PutBool(kv.Key, kv.Value.AsBool())
	case otellog.KindFloat64:
		into.PutDouble(kv.Key, kv.Value.AsFloat64())
	case otellog.KindMap:
		into.PutEmptyMap(kv.Key)
		intoMap, ok := into.Get(kv.Key)
		if !ok {
			return false
		}
		fromMap := kv.Value.AsMap()
		for _, subKv := range fromMap {
			if !mapAttributes(subKv, intoMap.Map()) {
				return false
			}
		}
	default:
		return false
	}
	return true
}
