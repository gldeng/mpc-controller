package logger

import (
	"fmt"
	sirupsenLogrus "github.com/sirupsen/logrus"
	"strings"
)

var _ Logger = (*logrus)(nil)

type logrus struct {
	l *sirupsenLogrus.Logger
}

func NewLogrus(l *sirupsenLogrus.Logger) Logger {
	return &logrus{l}
}

func (l *logrus) Debug(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Debug(msg)
}

func (l *logrus) Info(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Info(msg)
}

func (l *logrus) Warn(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Warn(msg)
}

func (l *logrus) Error(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Error(msg)
}

func (l *logrus) Fatal(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Fatal(msg)
}

func (l *logrus) With(fields ...Field) Logger {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl)
	return l
}

// ---

func (l *logrus) Debugf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.Debug(msg)
}

func (l *logrus) Infof(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.Info(msg)
}

func (l *logrus) Warnf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.Warn(msg)
}

func (l *logrus) Errorf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.Error(msg)
}

func (l *logrus) Fatalf(format string, a ...interface{}) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, a...), "\n")
	l.Fatalf(msg)
}
