package storage

import (
	"context"
	"encoding/hex"
	"github.com/avalido/mpc-controller/core"
)

var (
	_ core.Store = (*InMemoryDb)(nil)
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

func (i *InMemoryDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	bytes := i.data[hex.EncodeToString(key)]
	return bytes, nil
}

func (i *InMemoryDb) Exists(ctx context.Context, key []byte) (bool, error) {
	bytes, err := i.Get(ctx, key)
	return bytes != nil, err
}
