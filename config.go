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
		AddSource:   false,
		Level:       level,
		ReplaceAttr: c.fieldMapToPlaceKey(),
	}
	if c.Formatter.Format == FormatterGoogle {
		opts.ReplaceAttr = handlers.ReplaceAttrForGoogleFunc(c.fieldMapToPlaceKey())
	}
	var handler slog.Handler
	switch c.Formatter.Format {
	case FormatterText:
		handler = slog.NewTextHandler(os.Stderr, opts)
	case FormatterJSON, FormatterGoogle:
		handler = slog.NewJSONHandler(os.Stderr, opts)
	case "":
		logger.Warn("no slog format in config, using text handler")
	default:
		logger.Warn("unknown slog format in config, using text handler", "format", c.Formatter.Format)
	}
	if c.Formatter.Format == FormatterGoogle {
		handler = handlers.ForGoogleCloudLogging(handler, c.Formatter.Data)
	}
	if c.AddSource {
		// The order matters.
		// If AddCallerAndStack wraps the GoogleHandler, the caller field is added to the app context.
		handler = handlers.AddCallerAndStack(handler)
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
