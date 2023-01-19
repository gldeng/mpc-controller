package logger

// Field is the simple container for a single log field
type Field struct {
	Key   string
	Value interface{}
}

// Logger declares base logging methods
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	StdLogger
}

// StdLogger declares standard logging methods
type StdLogger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
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
