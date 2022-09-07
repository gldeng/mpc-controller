package badgerDB

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/dgraph-io/badger/v3"
)

func NewBadgerDBWithDefaultOptions(path string, logger logger.BadgerDBLogger) *badger.DB {
	opt := badger.DefaultOptions(path)
	opt.Logger = logger
	db, err := badger.Open(opt)
	if err != nil {
		panic("Failed to open badger database, error:" + err.Error())
	}
	return db
}
