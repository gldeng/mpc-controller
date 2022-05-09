package logger

import (
	uberZap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"time"
)

var DefaultLogger Logger

var DevMode bool

var once = new(sync.Once)

// Default return a Logger right depending on go.uber.org/zap Logger.
func Default() Logger {
	once.Do(func() {
		var logger *uberZap.Logger
		var logConfig uberZap.Config

		if DevMode {
			logConfig = uberZap.NewDevelopmentConfig()
			logConfig.EncoderConfig.EncodeTime = iso8601UTCTimeEncoder
			logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			logger, _ = logConfig.Build(uberZap.AddCallerSkip(2))
		} else {
			logConfig = uberZap.NewProductionConfig()
			logConfig.EncoderConfig.EncodeTime = iso8601UTCTimeEncoder
			logger, _ = logConfig.Build(uberZap.AddCallerSkip(2))
		}
		DefaultLogger = NewZap(logger)
	})
	return DefaultLogger
}

// A UTC variation of ZapCore.ISO8601TimeEncoder with millisecond precision
func iso8601UTCTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z"))
}

func Debug(msg string, fields ...Field) {
	logger := Default()
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	logger := Default()
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	logger := Default()
	logger.Warn(msg, fields...)
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
