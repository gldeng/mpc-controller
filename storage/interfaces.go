package storage

import (
	"context"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding low-level k-v db

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
	Set(ctx context.Context, key []byte, val interface{}) error
}

type UnmarshalGetter interface {
	Get(ctx context.Context, key []byte, val interface{}) error
}

type UnmarshalLister interface {
	List(ctx context.Context, prefix []byte, val interface{}) error
}
