package badgerDB

import (
	"github.com/dgraph-io/badger/v3"
)

func NewBadgerDBWithDefaultOptions(path string) *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		panic("Failed to open badger database")
	}
	return db
}
