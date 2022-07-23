package adapter

import (
	"fmt"
	"github.com/avalido/mpc-controller/logger"
)

var _ logger.BadgerDBLogger = (*BadgerDBLoggerAdapter)(nil)

type BadgerDBLoggerAdapter struct {
	logger.Logger
}

func (l *BadgerDBLoggerAdapter) Debugf(format string, a ...interface{}) {
	l.Debug(fmt.Sprintf(format, a...))
}

func (l *BadgerDBLoggerAdapter) Infof(format string, a ...interface{}) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l *BadgerDBLoggerAdapter) Warningf(format string, a ...interface{}) {
	l.Warn(fmt.Sprintf(format, a...))
}

func (l *BadgerDBLoggerAdapter) Errorf(format string, a ...interface{}) {
	l.Error(fmt.Sprintf(format, a...))
}
