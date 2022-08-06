package badgerDB

import (
	"context"
	"encoding/json"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"time"
)

var _ storage.KvDb = (*BadgerDB)(nil)

type BadgerDB struct {
	logger.Logger
	*badger.DB
}

// Classic k-v storage

func (b *BadgerDB) Set(ctx context.Context, key, val []byte) (err error) {
	err = backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
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
	err = backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
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
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get value by key. key:%v", string(key))
	return
}

func (b *BadgerDB) List(ctx context.Context, prefix []byte) ([][]byte, error) {
	return nil, errors.New("to to implemented") // todo
}

// With marshal and unmarshal support

func (b *BadgerDB) MSet(ctx context.Context, key []byte, val interface{}) error {
	valBytes, err := json.Marshal(val)
	if err != nil {
		return errors.WithStack(err)
	}

	err = b.Set(ctx, key, valBytes)

	return errors.WithStack(err)
}

func (b *BadgerDB) MGet(ctx context.Context, key []byte, val interface{}) error {
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

func (b *BadgerDB) MList(ctx context.Context, prefix []byte, val interface{}) error {
	return errors.New("to to implemented") // todo
}

// With Model(s) interface support

func (b *BadgerDB) SaveModel(ctx context.Context, data storage.Model) error {
	return errors.WithStack(b.MSet(ctx, data.Key(), data))
}

func (b *BadgerDB) LoadModel(ctx context.Context, data storage.Model) error {
	return errors.WithStack(b.MGet(ctx, data.Key(), data))
}

func (b *BadgerDB) ListModels(ctx context.Context, datum storage.Models) error {
	return errors.WithStack(b.MList(ctx, datum.Prefix(), datum))
}
