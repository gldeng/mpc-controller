package logger

import (
	"fmt"
	uberZap "go.uber.org/zap"
	"strings"
)

var _ Logger = (*zap)(nil)

type zap struct {
	l *uberZap.Logger
}

func NewZap(l *uberZap.Logger) Logger {
	return &zap{l}
}

func (l *zap) Debug(msg string, fields ...Field) {
	l.l.Debug(msg, l.zapFields(fields...)...)
}

func (l *zap) Info(msg string, fields ...Field) {
	l.l.Info(msg, l.zapFields(fields...)...)
}

func (l *zap) Warn(msg string, fields ...Field) {
	l.l.Warn(msg, l.zapFields(fields...)...)
}

func (l *zap) Error(msg string, fields ...Field) {
	l.l.Error(msg, l.zapFields(fields...)...)
}

func (l *zap) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, l.zapFields(fields...)...)
}

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

// ---

func (l *zap) Debugf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.l.Debug(msg)
}

func (l *zap) Infof(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.l.Info(msg)
}

func (l *zap) Warnf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.l.Warn(msg)
}

func (l *zap) Errorf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.l.Error(msg)
}

func (l *zap) Fatalf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.l.Fatal(msg)
}
