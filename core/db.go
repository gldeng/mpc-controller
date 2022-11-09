package core

import (
	"context"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding low-level k-v db

type SlimDb interface {
	Setter
	Getter
}

// Classic k-v storage

type Setter interface {
	Set(ctx context.Context, key, val []byte) error
}

type Getter interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
}
