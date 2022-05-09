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

// Info implements Logger.Debug for sirupsen/logrus logger
func (l *logrus) Info(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Info(msg)
}

// Warn implements Logger.Debug for sirupsen/logrus logger
func (l *logrus) Warn(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Warn(msg)
}

// Error implements Logger.Debug for sirupsen/logrus logger
func (l *logrus) Error(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Error(msg)
}

// Fatal implements Logger.Debug for sirupsen/logrus logger
func (l *logrus) Fatal(msg string, fields ...Field) {
	fieldsl := sirupsenLogrus.Fields{}
	for _, f := range fields {
		fieldsl[f.Key] = f.Value
	}
	l.l.WithFields(fieldsl).Fatal(msg)
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
