package logger

import (
	uberZap "go.uber.org/zap"
)

var _ Logger = (*zap)(nil)

type zap struct {
	l *uberZap.Logger
}

// NewZap instantiates new Logger using go.uber.org/zap
func NewZap(l *uberZap.Logger) Logger {
	return &zap{l}
}

// Debug implements Logger.Debug for go.uber.org/zap logger
func (l *zap) Debug(msg string, fields ...Field) {
	l.l.Debug(msg, l.zapFields(fields...)...)
}

// Info implements Logger.Info for go.uber.org/zap logger
func (l *zap) Info(msg string, fields ...Field) {
	l.l.Info(msg, l.zapFields(fields...)...)
}

// Warn implements Logger.Warn for go.uber.org/zap logger
func (l *zap) Warn(msg string, fields ...Field) {
	l.l.Warn(msg, l.zapFields(fields...)...)
}

// Error implements Logger.Error for go.uber.org/zap logger
func (l *zap) Error(msg string, fields ...Field) {
	l.l.Error(msg, l.zapFields(fields...)...)
}

// ErrorOnError implements Logger.Error for go.uber.org/zap logger
func (l *zap) ErrorOnError(err error, msg string, fields ...Field) {
	if err != nil {
		l.Error(msg, fields...)
	}
}

// Fatal implements Logger.Fatal for go.uber.org/zap logger
func (l *zap) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, l.zapFields(fields...)...)
}

// FatalOnError implements Logger.Fatal for go.uber.org/zap logger
func (l *zap) FatalOnError(err error, msg string, fields ...Field) {
	if err != nil {
		l.Fatal(msg, fields...)
	}
}

// With implements nested logger for go.uber.org/zap logger
func (l *zap) With(fields ...Field) Logger {
	return NewZap(l.l.With(l.zapFields(fields...)...))
}

func (l *zap) zapFields(fields ...Field) []uberZap.Field {
	result := make([]uberZap.Field, len(fields))
	for i, f := range fields {
		result[i] = uberZap.Any(f.Key, f.Value)
	}
	return result
}
