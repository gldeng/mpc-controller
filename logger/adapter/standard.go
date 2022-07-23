package adapter

import (
	"fmt"
	"github.com/avalido/mpc-controller/logger"
)

var _ logger.StdLogger = (*StdLoggerAdapter)(nil)

type StdLoggerAdapter struct {
	logger.Logger
}

func (l *StdLoggerAdapter) Debugf(format string, a ...interface{}) {
	l.Debug(fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) Infof(format string, a ...interface{}) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) Warnf(format string, a ...interface{}) {
	l.Warn(fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) WarnOnErrorf(err error, format string, a ...interface{}) {
	l.WarnOnError(err, fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) WarnOnNotOkf(ok bool, format string, a ...interface{}) {
	l.WarnOnNotOk(ok, fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) Errorf(format string, a ...interface{}) {
	l.Error(fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) ErrorOnErrorf(err error, format string, a ...interface{}) {
	l.ErrorOnError(err, fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) ErrorOnNotOkf(ok bool, format string, a ...interface{}) {
	l.ErrorOnNotOk(ok, fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) Fatalf(format string, a ...interface{}) {
	l.Fatalf(fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) FatalOnErrorf(err error, format string, a ...interface{}) {
	l.FatalOnError(err, fmt.Sprintf(format, a...))
}

func (l *StdLoggerAdapter) FatalOnNotOkf(ok bool, format string, a ...interface{}) {
	l.FatalOnNotOk(ok, fmt.Sprintf(format, a...))
}
