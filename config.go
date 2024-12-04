package logging

import (
	"github.com/zitadel/logging/handlers"
	"log/slog"
	"os"
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
	FormatterText   = "text"
	FormatterJSON   = "json"
	FormatterGoogle = "google"
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
		AddSource:   c.AddSource,
		Level:       level,
		ReplaceAttr: c.fieldMapToPlaceKey(),
	}

	switch c.Formatter.Format {
	case FormatterText:
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	case FormatterJSON:
		return slog.New(slog.NewJSONHandler(os.Stderr, opts))
	case FormatterGoogle:
		return slog.New(handlers.NewGoogle(os.Stderr, opts, c.Formatter.Data))
	case "":
		logger.Warn("no slog format in config, using text handler")
	default:
		logger.Warn("unknown slog format in config, using text handler", "format", c.Formatter.Format)
	}
	return slog.New(slog.NewTextHandler(os.Stderr, opts))
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
