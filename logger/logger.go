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
	WarnOnError(err error, msg string, fields ...Field)
	WarnOnNotOk(ok bool, msg string, fields ...Field)

	Error(msg string, fields ...Field)
	ErrorOnError(err error, msg string, fields ...Field)
	ErrorOnNotOk(ok bool, msg string, fields ...Field)

	Fatal(msg string, fields ...Field)
	FatalOnError(err error, msg string, fields ...Field)
	FatalOnNotOk(ok bool, msg string, fields ...Field)

	With(fields ...Field) Logger
}

// StdLogger declares standard logging methods
type StdLogger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})

	Warnf(string, ...interface{})
	WarnOnErrorf(error, string, ...interface{})
	WarnOnNotOkf(bool, string, ...interface{})

	Errorf(string, ...interface{})
	ErrorOnErrorf(error, string, ...interface{})
	ErrorOnNotOkf(bool, string, ...interface{})

	Fatalf(string, ...interface{})
	FatalOnErrorf(error, string, ...interface{})
	FatalOnNotOkf(bool, string, ...interface{})
}
