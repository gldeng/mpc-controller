package logger

import (
	uberZap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

var DefaultLogger Logger

var DevMode bool

// Default return a Logger right depending on go.uber.org/zap Logger.
func Default() Logger {
	var logger *uberZap.Logger
	var logConfig uberZap.Config

	if DevMode {
		logConfig = uberZap.NewDevelopmentConfig()
		logConfig.EncoderConfig.EncodeTime = iso8601LocalTimeEncoder
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = logConfig.Build(uberZap.AddCallerSkip(1))
	} else {
		logConfig = uberZap.NewProductionConfig()
		logConfig.EncoderConfig.EncodeTime = iso8601LocalTimeEncoder
		logger, _ = logConfig.Build(uberZap.AddCallerSkip(1))
	}
	DefaultLogger = NewZap(logger)
	return DefaultLogger
}

// A UTC variation of ZapCore.ISO8601TimeEncoder with nanosecond precision
func iso8601UTCTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000000000Z"))
}

// A Local variation of ZapCore.ISO8601TimeEncoder with nanosecond precision
func iso8601LocalTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Local().Format("2006-01-02T15:04:05.000000000Z"))
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

func WarnOnError(err error, msg string, fields ...Field) {
	logger := Default()
	logger.WarnOnError(err, msg, fields...)
}

func WarnOnNotOk(ok bool, msg string, fields ...Field) {
	logger := Default()
	logger.WarnOnNotOk(ok, msg, fields...)
}

func Error(msg string, fields ...Field) {
	logger := Default()
	logger.Error(msg, fields...)
}

func ErrorOnError(err error, msg string, fields ...Field) {
	logger := Default()
	logger.ErrorOnError(err, msg, fields...)
}

func ErrorOnNotOk(ok bool, msg string, fields ...Field) {
	logger := Default()
	logger.ErrorOnNotOk(ok, msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	logger := Default()
	logger.Fatal(msg, fields...)
}

func FatalOnError(err error, msg string, fields ...Field) {
	logger := Default()
	logger.FatalOnError(err, msg, fields...)
}

func FatalOnNotOk(ok bool, msg string, fields ...Field) {
	logger := Default()
	logger.FatalOnNotOk(ok, msg, fields...)
}

func With(fields ...Field) Logger {
	logger := Default()
	logger.With(fields...)
	return logger
}
