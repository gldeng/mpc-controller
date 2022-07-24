package logger

import "fmt"

// Field is the simple container for a single log field
type Field struct {
	Key   string
	Value interface{}
}

func AppendErrorFiled(err error, fields ...Field) []Field {
	var errorFields []Field
	errorFields = append(errorFields, fields...)
	errMsg := fmt.Sprintf("%+v", err)
	errorFields = append(errorFields, Field{"error", errMsg})
	return errorFields
}

// Logger declares base logging methods
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)

	Warn(msg string, fields ...Field)
	WarnOnError(err error, msg string, fields ...Field)
	WarnOnTrue(ok bool, msg string, fields ...Field)

	Error(msg string, fields ...Field)
	ErrorOnError(err error, msg string, fields ...Field)
	ErrorOnTrue(ok bool, msg string, fields ...Field)

	Fatal(msg string, fields ...Field)
	FatalOnError(err error, msg string, fields ...Field)
	FatalOnTrue(ok bool, msg string, fields ...Field)

	With(fields ...Field) Logger
}

// StdLogger declares standard logging methods
type StdLogger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})

	Warnf(string, ...interface{})
	WarnOnErrorf(error, string, ...interface{})
	WarnOnTruef(bool, string, ...interface{})

	Errorf(string, ...interface{})
	ErrorOnErrorf(error, string, ...interface{})
	ErrorOnTruef(bool, string, ...interface{})

	Fatalf(string, ...interface{})
	FatalOnErrorf(error, string, ...interface{})
	FatalOnTruef(bool, string, ...interface{})
}

// BadgerDBLogger declares customized logging methods for BadgerDB
// as defined in https://github.com/dgraph-io/badger/blob/master/logger.go.
// Copy it here just for clarity.
type BadgerDBLogger interface {
	Errorf(string, ...interface{})
	Warningf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}
