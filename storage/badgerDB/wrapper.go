package badgerDB

import (
	"context"
	"encoding/json"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

type BadgerDB struct {
	logger.Logger
	*badger.DB
}

// Classic k-v storage

func (b *BadgerDB) Set(ctx context.Context, key, val []byte) (err error) {
	err = backoff.RetryFnExponentialForever(b.Logger, ctx, func() error {
		err = b.Update(func(txn *badger.Txn) error {
			err := txn.Set(key, val)
			return errors.WithStack(err)
		})
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	return errors.WithStack(err)
}

func (b *BadgerDB) Get(ctx context.Context, key []byte) (value []byte, err error) {
	err = backoff.RetryFnExponentialForever(b.Logger, ctx, func() error {
		err = b.View(func(txn *badger.Txn) error {
			item, err := txn.Get(key)
			if err != nil {
				return errors.WithStack(err)
			}

			err = item.Value(func(val []byte) error {
				value = val
				return nil
			})
			return errors.WithStack(err)
		})
		return errors.WithStack(err)
	})
	err = errors.WithStack(err)
	return
}

func (b *BadgerDB) List(ctx context.Context, prefix []byte) ([][]byte, error) {
	return nil, errors.New("to to implemented") // todo
}

// With marshal and unmarshal support

func (b *BadgerDB) MarshalSet(ctx context.Context, key []byte, val interface{}) error {
	valBytes, err := json.Marshal(val)
	if err != nil {
		return errors.WithStack(err)
	}

	err = b.Set(ctx, key, valBytes)

	return errors.WithStack(err)
}

func (b *BadgerDB) MarshalGet(ctx context.Context, key []byte, val interface{}) error {
	valBytes, err := b.Get(ctx, key)
	if err != nil {
		return errors.WithStack(err)
	}

	err = json.Unmarshal(valBytes, val)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (b *BadgerDB) MarshalList(ctx context.Context, prefix []byte, val interface{}) error {
	return errors.New("to to implemented") // todo
}
