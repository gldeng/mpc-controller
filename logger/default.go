package logger

import (
	uberZap "go.uber.org/zap"
	"sync"
)

var DefaultLogger Logger

var once = new(sync.Once)

// Default return a Logger right depending on go.uber.org/zap Logger.
func Default() Logger {
	once.Do(func() {
		if DefaultLogger == nil {
			logger, _ := uberZap.NewProduction()
			DefaultLogger = NewZap(logger)
		}
	})
	return DefaultLogger
}

func Debug(msg string, fields ...Field) {
	logger := Default()
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	logger := Default()
	logger.Info(msg, fields...)
}

func Error(msg string, fields ...Field) {
	logger := Default()
	logger.Error(msg, fields...)
}

func With(fields ...Field) Logger {
	logger := Default()
	logger.With(fields...)
	return logger
}
