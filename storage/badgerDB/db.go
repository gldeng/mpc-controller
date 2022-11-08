package badgerDB

import (
	"github.com/dgraph-io/badger/v3"
)

func NewBadgerDBWithDefaultOptions(path string, logger badger.Logger) *badger.DB {
	opt := badger.DefaultOptions(path)
	opt.Logger = logger
	db, err := badger.Open(opt)
	if err != nil {
		panic("Failed to open badger database, error:" + err.Error())
	}
	return db
}
