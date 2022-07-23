package badgerDB

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/logger/adapter"
	"github.com/dgraph-io/badger/v3"
)

func NewBadgerDBWithDefaultOptions(path string, logger logger.Logger) *badger.DB {
	opt := badger.DefaultOptions(path)
	opt.Logger = &adapter.BadgerDBLoggerAdapter{StdLogger: &adapter.StdLoggerAdapter{Logger: logger}}
	db, err := badger.Open(opt)
	if err != nil {
		panic("Failed to open badger database")
	}
	return db
}
