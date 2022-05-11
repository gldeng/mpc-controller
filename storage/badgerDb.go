package storage

import (
	"encoding/json"
	"github.com/avalido/mpc-controller/logger"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

var _ = (*badgerDb)(nil)

type badgerDb struct {
	logger.Logger
	db *badger.DB
}

type Config interface {
}

// todo: more configs tuned
// note: pay attention to transaction usage.
// note: pay attention to stream feature usage.
// reference: https://dgraph.io/docs/badger/get-started/

func New(log logger.Logger, path string) Storer {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	db, err := badger.Open(badger.DefaultOptions(path))
	log.FatalOnError(err, "Failed to open badger database", logger.Field{"error", err})

	// Your code here…
	return &badgerDb{log, db}
}

func (b *badgerDb) StoreGroupInfo(g *GroupInfo) error {
	bytes, err := json.Marshal(g)
	if err != nil {
		b.Error("Failed to marshal", []logger.Field{{"groupInfo", g}, {"error", err}}...)
		return errors.Wrapf(err, "failed to marshal: %+v", g)
	}
	key := prefixGroupInfo + "-" + g.GroupIdHex
	err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), bytes)
		return errors.WithStack(err)
	})
	if err != nil {
		b.Error("Failed to store", []logger.Field{{"groupInfo", g}, {"error", err}}...)
		return errors.Wrapf(err, "failed to store %+v", g)
	}

	b.Info("Group info stored", []logger.Field{{"key", key}, {"value", g}}...)
	return nil
}

func (b *badgerDb) LoadGroupInfo(groupIdHex string) (*GroupInfo, error) {
	var g GroupInfo

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixGroupInfo + "-" + groupIdHex))
		if err != nil {
			return errors.WithStack(err)
		}

		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, &g)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal")
			}
			return nil
		})
		return errors.WithStack(err)
	})

	if err != nil {
		b.Error("Failed to load group info", []logger.Field{{"groupIdHex", groupIdHex}, {"error", err}}...)
		return nil, errors.Wrapf(err, "failed to load group info, groupIdHex: %q", groupIdHex)
	}

	return &g, nil
}

func (b *badgerDb) LoadGroupInfos() ([]*GroupInfo, error) {
	var gs []*GroupInfo
	b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefixGroupInfo)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var g GroupInfo
				err := json.Unmarshal(v, &g)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal")
				}
				gs = append(gs, &g)
				return nil
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})

	return gs, nil
}

func (b *badgerDb) StoreParticipantInfo(p *ParticipantInfo) error {
	bytes, err := json.Marshal(p)
	if err != nil {
		b.Error("Failed to marshal", []logger.Field{{"participantInfo", p}, {"error", err}}...)
		return errors.Wrapf(err, "failed to marshal %+v", p)
	}
	key := prefixParticipantInfo + "-" + p.PubKeyHashHex + "-" + p.GroupIdHex
	err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), bytes)
		return errors.WithStack(err)
	})
	if err != nil {
		b.Error("Failed to store", []logger.Field{{"participantInfo", p}, {"error", err}}...)
		return errors.Wrapf(err, "failed to store %+v", p)
	}

	b.Info("participant info stored", []logger.Field{{"key", key}, {"value", p}}...)
	return nil
}

func (b *badgerDb) LoadParticipantInfo(pubKeyHashHex, groupId string) (*ParticipantInfo, error) {
	var v ParticipantInfo

	key := prefixParticipantInfo + "-" + pubKeyHashHex + "-" + groupId
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return errors.WithStack(err)
		}

		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, &v)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal")
			}
			return nil
		})
		return errors.WithStack(err)
	})

	if err != nil {
		b.Error("Failed to load participant info", []logger.Field{{"pubKeyHashHex-groupId", key}, {"error", err}}...)
		return nil, errors.Wrapf(err, "failed to load participant info, pubKeyHashHex-groupId: %q", key)
	}

	return &v, nil
}

func (b *badgerDb) LoadParticipantInfos(pubKeyHashHex string) ([]*ParticipantInfo, error) {
	var ps []*ParticipantInfo
	b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefixParticipantInfo + "-" + pubKeyHashHex)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var p ParticipantInfo
				err := json.Unmarshal(v, &p)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal")
				}
				ps = append(ps, &p)
				return nil
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})

	return ps, nil
}

//

func (b *badgerDb) StoreGeneratedPubKeyInfo(pk *GeneratedPubKeyInfo) error {
	bytes, err := json.Marshal(pk)
	if err != nil {
		b.Error("Failed to marshal", []logger.Field{{"generatedPubKeyInfo", pk}, {"error", err}}...)
		return errors.Wrapf(err, "failed to marshal %+v", pk)
	}
	key := prefixGeneratedPubKeyInfo + "-" + pk.PubKeyHashHex
	err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), bytes)
		return errors.WithStack(err)
	})
	if err != nil {
		b.Error("Failed to store", []logger.Field{{"generatedPubKeyInfo", pk}, {"error", err}}...)
		return errors.Wrapf(err, "failed to store %+v", pk)
	}

	b.Info("generated public key info stored", []logger.Field{{"key", key}, {"value", pk}}...)
	return nil
}

func (b *badgerDb) LoadGeneratedPubKeyInfo(pubKeyHashHex string) (*GeneratedPubKeyInfo, error) {
	var g GeneratedPubKeyInfo

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixGeneratedPubKeyInfo + "-" + pubKeyHashHex))
		if err != nil {
			return errors.WithStack(err)
		}

		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, &g)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal")
			}
			return nil
		})
		return errors.WithStack(err)
	})

	if err != nil {
		b.Error("Failed to load generated public key info", []logger.Field{{"pubKeyHashHex", pubKeyHashHex}, {"error", err}}...)
		return nil, errors.Wrapf(err, "failed to load generated public key info, pubKeyHashHex: %q", pubKeyHashHex)
	}

	return &g, nil
}

func (b *badgerDb) LoadGeneratedPubKeyInfos(groupIdHexs []string) ([]*GeneratedPubKeyInfo, error) {
	var ps []*GeneratedPubKeyInfo

	var groupIdHexMap = make(map[string]bool)
	for _, v := range groupIdHexs {
		groupIdHexMap[v] = true
	}

	b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefixGeneratedPubKeyInfo)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var p GeneratedPubKeyInfo
				err := json.Unmarshal(v, &p)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal")
				}
				_, ok := groupIdHexMap[p.GroupIdHex]
				if ok {
					ps = append(ps, &p)

				}
				return nil
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})

	return ps, nil
}

//

func (b *badgerDb) StoreKeygenRequestInfo(k *KeygenRequestInfo) error {
	bytes, err := json.Marshal(k)
	if err != nil {
		b.Error("Failed to marshal", []logger.Field{{"keygenRequestInfo", k}, {"error", err}}...)
		return errors.Wrapf(err, "failed to marshal %+v", k)
	}
	key := prefixKeygenRequestInfo + "-" + k.RequestIdHex
	err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), bytes)
		return errors.WithStack(err)
	})
	if err != nil {
		b.Error("Failed to store", []logger.Field{{"keygenRequestInfo", k}, {"error", err}}...)
		return errors.Wrapf(err, "failed to store %+v", k)
	}

	b.Info("keygen request info stored", []logger.Field{{"key", key}, {"value", k}}...)
	return nil
}

func (b *badgerDb) LoadKeygenRequestInfo(reqIdHex string) (*KeygenRequestInfo, error) {
	var v KeygenRequestInfo

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixKeygenRequestInfo + "-" + reqIdHex))
		if err != nil {
			return errors.WithStack(err)
		}

		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, &v)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal")
			}
			return nil
		})
		return errors.WithStack(err)
	})

	if err != nil {
		b.Error("Failed to load keygen request info", []logger.Field{{"reqIdHex", reqIdHex}, {"error", err}}...)
		return nil, errors.Wrapf(err, "failed to load keygen request info: %q", reqIdHex)
	}

	return &v, nil
}

//

func (b *badgerDb) Close() error {
	return b.db.Close()
}
