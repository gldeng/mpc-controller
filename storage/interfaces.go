package storage

import (
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
	MarshalSet(ctx context.Context, key []byte, val interface{}) error
}

type UnmarshalGetter interface {
	MarshalGet(ctx context.Context, key []byte, val interface{}) error
}

type UnmarshalLister interface {
	MarshalList(ctx context.Context, prefix []byte, val interface{}) error
}
