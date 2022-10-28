package storage

import (
	"context"
	"encoding/hex"
)

type InMemoryDb struct {
	data map[string][]byte
}

func NewInMemoryDb() *InMemoryDb {
	return &InMemoryDb{data: map[string][]byte{}}
}

func (i *InMemoryDb) Set(ctx context.Context, key, val []byte) error {
	i.data[hex.EncodeToString(key)] = val
	return nil
}

func (i InMemoryDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	bytes := i.data[hex.EncodeToString(key)]
	return bytes, nil
}
