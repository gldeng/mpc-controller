package adapter

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/dgraph-io/badger/v3"
)

var _ badger.Logger = (*BadgerDBLoggerAdapter)(nil)

type BadgerDBLoggerAdapter struct {
	logger.StdLogger
}

func (l *BadgerDBLoggerAdapter) Warningf(format string, a ...interface{}) {
	l.Warnf(format, a...)
}
