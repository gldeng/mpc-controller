package core

import (
	"context"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding low-level k-v db

type Store interface {
	Set(ctx context.Context, key, val []byte) error
	Get(ctx context.Context, key []byte) ([]byte, error)
}
