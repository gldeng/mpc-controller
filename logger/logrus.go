package logger

import (
	sirupsenLogrus "github.com/sirupsen/logrus"
)

var _ Logger = (*logrus)(nil)

type logrus struct {
	l *sirupsenLogrus.Logger
}

// NewLogrus instantiates new Logger using using sirupsen/logrus
func NewLogrus(l *sirupsenLogrus.Logger) Logger {
	return &logrus{l}
}

// Debug implements Logger.Debug for sirupsen/logrus logger
func (l *logrus) Debug(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Debug(msg)
}

// Info implements Logger.Info for sirupsen/logrus logger
func (l *logrus) Info(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Info(msg)
}

// Warn implements Logger.Warn for sirupsen/logrus logger
func (l *logrus) Warn(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Warn(msg)
}

// Error implements Logger.Error for sirupsen/logrus logger
func (l *logrus) Error(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Error(msg)
}

// ErrorOnError implements Logger.Error for sirupsen/logrus logger
func (l *logrus) ErrorOnError(err error, msg string, fields ...Field) {
	if err != nil {
		l.Error(msg, fields...)
	}
}

// Fatal implements Logger.Fatal for sirupsen/logrus logger
func (l *logrus) Fatal(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Fatal(msg)
}

// FatalOnError implements Logger.Fatal for sirupsen/logrus logger
func (l *logrus) FatalOnError(err error, msg string, fields ...Field) {
	if err != nil {
		l.Fatal(msg, fields...)
	}
}

// With implements nested logrus for sirupsen/logrus logger
func (l *logrus) With(fields ...Field) Logger {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl)
	return l
}
