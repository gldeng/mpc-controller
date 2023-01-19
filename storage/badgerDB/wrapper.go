package badgerDB

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"time"
)

var _ core.Store = (*BadgerDB)(nil)

type BadgerDB struct {
	logger.Logger
	*badger.DB
}

// Classic k-v storage

func (b *BadgerDB) Set(ctx context.Context, key, val []byte) (err error) {
	err = backoff.RetryFnExponentialForever(b.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		err = b.Update(func(txn *badger.Txn) error {
			err := txn.Set(key, val)
			return errors.WithStack(err)
		})
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to set k-v. key:%v, val:%v", string(key), string(val))
	return
}

func (b *BadgerDB) Get(ctx context.Context, key []byte) (value []byte, err error) {
	err = backoff.RetryFnExponential10Times(b.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
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
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return false, errors.WithStack(err)
			}
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get value by key. key:%v", string(key))
	return
}

func (b *BadgerDB) Exists(ctx context.Context, key []byte) (found bool, err error) {
	err = backoff.RetryFnExponential10Times(b.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		err = b.View(func(txn *badger.Txn) error {
			_, err := txn.Get(key)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		})
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return false, nil
			}
			return true, errors.WithStack(err)
		}
		found = true
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get value by key. key:%v", string(key))
	return
}
