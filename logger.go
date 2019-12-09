package logging

import (
	"encoding/json"
	"io"

	"github.com/sirupsen/logrus"
)

type Logger logrus.Logger

var log *Logger = (*Logger)(logrus.StandardLogger())

func SetOutput(out io.Writer) {
	(*logrus.Logger)(log).SetOutput(out)
}

func SetFormatter(formatter logrus.Formatter) {
	(*logrus.Logger)(log).SetFormatter(formatter)
}
func SetLevel(level logrus.Level) {
	(*logrus.Logger)(log).SetLevel(level)
}
func SetReportCaller(include bool) {
	(*logrus.Logger)(log).SetReportCaller(include)
}

func (l *Logger) UnmarshalJSON(data []byte) error {
	log = (*Logger)(logrus.New())
	c := config{}
	err := json.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	err = c.parseFormatter()
	if err != nil {
		return err
	}

	err = c.parseLevel()
	if err != nil {
		return err
	}

	err = c.unmarshallFormatter()
	if err != nil {
		return err
	}

	log.ReportCaller = c.LogCaller

	c.setGlobal()

	return nil
}
