package storage

import (
	"context"
	"encoding/json"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
)

var _ Storer = (*badgerDb)(nil)

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

func (b *badgerDb) StoreGroupInfo(ctx context.Context, g *GroupInfo) error {
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

	return nil
}

func (b *badgerDb) LoadGroupInfo(ctx context.Context, groupIdHex string) (*GroupInfo, error) {
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

func (b *badgerDb) LoadGroupInfos(ctx context.Context) ([]*GroupInfo, error) {
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

func (b *badgerDb) StoreParticipantInfo(ctx context.Context, p *ParticipantInfo) error {
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

	return nil
}

func (b *badgerDb) LoadParticipantInfo(ctx context.Context, pubKeyHashHex, groupId string) (*ParticipantInfo, error) {
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

func (b *badgerDb) LoadParticipantInfos(ctx context.Context, pubKeyHashHex string) ([]*ParticipantInfo, error) {
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

func (b *badgerDb) GetIndex(ctx context.Context, partiPubKeyHashHex, genPubKeyHashHex string) (*big.Int, error) {
	genPubKeyInfo, err := b.LoadGeneratedPubKeyInfo(ctx, genPubKeyHashHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	partInfo, err := b.LoadParticipantInfo(ctx, partiPubKeyHashHex, genPubKeyInfo.GroupIdHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return big.NewInt(int64(partInfo.Index)), nil
}

func (b *badgerDb) GetGroupIds(ctx context.Context, partiPubKeyHashHex string) ([][32]byte, error) {
	partInfos, err := b.LoadParticipantInfos(ctx, partiPubKeyHashHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var groupIds [][32]byte

	for _, partInfo := range partInfos {
		var groupId [32]byte
		groupIdRaw := common.Hex2BytesFixed(partInfo.GroupIdHex, 32)
		copy(groupId[:], groupIdRaw)
		groupIds = append(groupIds, groupId)
	}

	return groupIds, nil
}

func (b *badgerDb) GetPubKeys(ctx context.Context, partiPubKeyHashHex string) ([][]byte, error) {
	partyInfos, err := b.LoadParticipantInfos(ctx, partiPubKeyHashHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var groupIdHexs []string
	for _, partyInfo := range partyInfos {
		groupIdHexs = append(groupIdHexs, partyInfo.GroupIdHex)
	}
	if len(groupIdHexs) == 0 {
		return nil, errors.New("found no group")
	}

	genPubKeyInfos, err := b.LoadGeneratedPubKeyInfos(ctx, groupIdHexs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(genPubKeyInfos) == 0 {
		return nil, errors.New("found no generated public key")
	}

	var genPubKeyBytes [][]byte
	for _, genPubKeyInfo := range genPubKeyInfos {
		dnmGenPubKeyBytes, err := crypto.DenormalizePubKeyFromHex(genPubKeyInfo.PubKeyHex) // for Ethereum compatibility
		if err != nil {
			return nil, errors.WithStack(err)
		}
		genPubKeyBytes = append(genPubKeyBytes, dnmGenPubKeyBytes)
	}

	return genPubKeyBytes, nil
}

//

func (b *badgerDb) StoreGeneratedPubKeyInfo(ctx context.Context, pk *GeneratedPubKeyInfo) error {
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

	return nil
}

func (b *badgerDb) LoadGeneratedPubKeyInfo(ctx context.Context, pubKeyHashHex string) (*GeneratedPubKeyInfo, error) {
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

func (b *badgerDb) LoadGeneratedPubKeyInfos(ctx context.Context, groupIdHexs []string) ([]*GeneratedPubKeyInfo, error) {
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

func (b *badgerDb) GetPariticipantKeys(ctx context.Context, genPubKeyHashHex string, indices []*big.Int) ([]string, error) {
	genPkInfo, err := b.LoadGeneratedPubKeyInfo(ctx, genPubKeyHashHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	groupInfo, err := b.LoadGroupInfo(ctx, genPkInfo.GroupIdHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var partPubKeyHexs []string
	for _, ind := range indices {
		partPubKeyHex := groupInfo.PartPubKeyHexs[ind.Uint64()-1]
		partPubKeyHexs = append(partPubKeyHexs, partPubKeyHex)
	}
	return partPubKeyHexs, nil
}

//

func (b *badgerDb) StoreKeygenRequestInfo(ctx context.Context, k *KeygenRequestInfo) error {
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

	return nil
}

func (b *badgerDb) LoadKeygenRequestInfo(ctx context.Context, reqIdHex string) (*KeygenRequestInfo, error) {
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
