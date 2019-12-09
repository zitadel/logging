package logging

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type config struct {
	Level       string    `json:"level"`
	Formatter   formatter `json:"formatter"`
	LocalLogger bool      `json:"localLogger"`
	LogCaller   bool      `json:"logCaller"`
}

type formatter struct {
	Format string                 `json:"format"`
	Data   map[string]interface{} `json:"data"`
}

func (c *config) setGlobal() {
	if c.LocalLogger {
		return
	}
	logrus.SetFormatter(log.Formatter)
	logrus.SetLevel(log.Level)
	logrus.SetReportCaller(log.ReportCaller)
	log = (*Logger)(logrus.StandardLogger())
}

func (c *config) unmarshallFormatter() error {
	formatterData, err := json.Marshal(c.Formatter.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(formatterData, log.Formatter)
}

func (c *config) parseLevel() error {
	if c.Level == "" {
		log.Level = logrus.InfoLevel
		return nil
	}

	level, err := logrus.ParseLevel(c.Level)
	if err != nil {
		return err
	}
	log.Level = level

	return nil
}

func (c *config) parseFormatter() error {
	switch c.Formatter.Format {
	case "json":
		log.Formatter = &logrus.JSONFormatter{}
	case "text", "":
		log.Formatter = &logrus.TextFormatter{}
	default:
		return fmt.Errorf("%s formatter not supported", c.Formatter)
	}
	return nil
}
