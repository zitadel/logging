package logging

import (
	"log/slog"
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/zitadel/logging/records/v1"
)

type Config struct {
	Level       string    `json:"level"`
	Formatter   formatter `json:"formatter"`
	LocalLogger bool      `json:"localLogger"`
	AddSource   bool      `json:"addSource"`
}

type formatter struct {
	Format string                 `json:"format"`
	Data   map[string]interface{} `json:"data"`
}

const (
	FormatterText    = "text"
	FormatterJSON    = "json"
	FormatterZitadel = "zitadel"
)

// Slog constructs a slog.Logger with the Formatter and Level from config.
func (c *Config) Slog() *slog.Logger {
	logger := slog.Default()

	var level slog.Level
	if err := level.UnmarshalText([]byte(c.Level)); err != nil {
		logger.Warn("invalid config, using default slog", "err", err)
		return logger
	}
	opts := &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: c.fieldMapToPlaceKey(),
	}
	var handler slog.Handler
	switch c.Formatter.Format {
	case FormatterText:
		handler = slog.NewTextHandler(os.Stderr, opts)
	case FormatterJSON:
		handler = slog.NewJSONHandler(os.Stderr, opts)
	case FormatterZitadel:
		handler = record_v1.NewZitadelHandler(os.Stderr, opts, func(record *record_v1.Record) proto.Message {
			return &VersionedRecord{Record: &VersionedRecord_RecordV1{RecordV1: record}}
		}, "my service", "my version", "my pod id", map[string]any{"region": "AU1"})
	case "":
		logger.Warn("no slog format in config, using text handler")
	default:
		logger.Warn("unknown slog format in config, using text handler", "format", c.Formatter.Format)
	}
	return slog.New(handler)
}

func (c *Config) fieldMapToPlaceKey() func(groups []string, a slog.Attr) slog.Attr {
	fieldMap, ok := c.Formatter.Data["fieldmap"].(map[string]interface{})
	if !ok {
		return nil
	}
	return func(groups []string, a slog.Attr) slog.Attr {
		for key, newKey := range fieldMap {
			if a.Key == key {
				a.Key = newKey.(string)
				return a
			}
		}
		return a
	}
}
