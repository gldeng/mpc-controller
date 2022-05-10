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
	ErrorOnError(err error, msg string, fields ...Field)

	Fatal(msg string, fields ...Field)
	FatalOnError(err error, msg string, fields ...Field)

	With(fields ...Field) Logger
}
