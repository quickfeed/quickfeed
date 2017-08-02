package logger

import (
	"github.com/sirupsen/logrus"
)

// IgnoreFieldsFormatter ignores the specified fields before writing to the log.
type IgnoreFieldsFormatter struct {
	logrus.Formatter
	Ignore []string
}

// Format implements the logrus.Formatter interface.
func (f IgnoreFieldsFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	for _, field := range f.Ignore {
		delete(entry.Data, field)
	}

	return f.Formatter.Format(entry)
}

// NewDevFormatter returns a new development log formatter.
func NewDevFormatter(formatter logrus.Formatter) IgnoreFieldsFormatter {
	return IgnoreFieldsFormatter{
		Formatter: formatter,
		Ignore:    []string{"latency", "latency_human", "bytes_in", "bytes_out", "referer"},
	}
}
