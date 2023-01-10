package resiliency

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger() *StandardLogger {
	baseLogger := logrus.New()
	baseLogger.SetOutput(os.Stdout)
	baseLogger.SetLevel(logrus.InfoLevel)
	var sl = &StandardLogger{baseLogger}
	sl.Formatter = &logrus.JSONFormatter{}
	return sl
}

var (
	internalError = LogEvent{InternalError, "Internal Error: %v"}
	LogInfo       = LogEvent{Info, "Info : %v"}
	LogDebug      = LogEvent{Info, "Info : %v"}
)

func (l *StandardLogger) LogInternalError(message string) {
	l.Errorf(internalError.message, message)
}

func (l *StandardLogger) LogInfo(message string) {
	l.Infof(LogInfo.message, message)
}

func (l *StandardLogger) LogDebug(message string) {
	l.Debugf(LogDebug.message, message)
}
