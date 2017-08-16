package util

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

const (
	// HappoAgentLogLevelDebug shows LogLevel debug
	HappoAgentLogLevelDebug = "debug"
	// HappoAgentLogLevelInfo shows LogLevel info
	HappoAgentLogLevelInfo = "info"
	// HappoAgentLogLevelWarn shows LogLevel warn
	HappoAgentLogLevelWarn = "warn"
	// HappoAgentLogLevelDefault shows default LogLevel
	HappoAgentLogLevelDefault = HappoAgentLogLevelWarn
)

var logger *logrus.Logger

// HappoAgentFormatter log formatter for happo-agent
type HappoAgentFormatter struct {
}

func init() {
	logger = logrus.New()
	logger.Out = os.Stderr

	logger.Formatter = new(HappoAgentFormatter)
}

// Format implements HappoAgentFormatter
// 2006-01-02 15:04:05 [LogLevel] message key=value,key=value...
func (f *HappoAgentFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	b.WriteString(fmt.Sprintf("%s [%s] %s",
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Level.String(),
		strings.Trim(entry.Message, "\n"),
	))

	if len(entry.Data) > 0 {
		var valueMessage string
		fields := make([]string, 0, len(entry.Data))
		for key, value := range entry.Data {
			switch value := value.(type) {
			case string:
				valueMessage = value
			case error:
				valueMessage = value.Error()
			default:
				valueMessage = fmt.Sprint(value)
			}
			fields = append(fields,
				fmt.Sprintf("%s=%s",
					strings.Trim(key, " "),
					strings.Trim(valueMessage, " "),
				))
		}
		b.WriteString(fmt.Sprintf(" %s", strings.Join(fields, ",")))
	}

	b.WriteByte('\n')
	return b.Bytes(), nil

}

// HappoAgentLogger returns custom logger
func HappoAgentLogger() *logrus.Logger {
	return logger
}

// SetLogLevel parse string and set log level
func SetLogLevel(logLevel string) {
	logLevel = strings.ToLower(logLevel)
	logLevel = strings.TrimSpace(logLevel)
	switch logLevel {
	case HappoAgentLogLevelInfo:
		logger.Level = logrus.InfoLevel
	case HappoAgentLogLevelDebug:
		logger.Level = logrus.DebugLevel
	default:
		logger.Level = logrus.WarnLevel
	}
	logger.WithField("logger.Level", logger.Level.String()).Debug("set LogLevel")
}
