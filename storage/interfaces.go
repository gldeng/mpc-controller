package storage

import (
	"bytes"
	"context"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding low-level k-v db

type KvDb interface {
	Setter
	Getter
	Lister

	MarshalSetter
	UnmarshalGetter
	UnmarshalLister

	ModelSaver
	ModelLoader
	UnmarshalLister
}

// Classic k-v storage

type Setter interface {
	Set(ctx context.Context, key, val []byte) error
}

type Getter interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
}

type Lister interface {
	List(ctx context.Context, prefix []byte) ([][]byte, error)
}

// With marshal and unmarshal support

type MarshalSetter interface {
	MSet(ctx context.Context, key []byte, val interface{}) error
}

type UnmarshalGetter interface {
	MGet(ctx context.Context, key []byte, val interface{}) error
}

type UnmarshalLister interface {
	MList(ctx context.Context, prefix []byte, val interface{}) error
}

// With Model(s) interface support

type Model interface {
	Key() []byte
}

type Models interface {
	Prefix() []byte
}

type ModelSaver interface {
	SaveModel(ctx context.Context, data Model) error
}

type ModelLoader interface {
	LoadModel(ctx context.Context, data Model) error
}

type ModelLister interface {
	ListModels(ctx context.Context, datum Models) error
}

// Handy function to concat a key

const (
	KeyPayLoadLength = 32
)

type KeyPrefix []byte
type KeyPayload [KeyPayLoadLength]byte

func Key(pre KeyPrefix, payload KeyPayload) []byte {
	return JoinWithHyphen([][]byte{pre, payload[:]})
}

func JoinWithHyphen(s [][]byte) []byte {
	return bytes.Join(s, []byte("-"))
}
