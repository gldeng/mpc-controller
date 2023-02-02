package common

import (
	"context"
	"encoding/json"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/utils/crypto"
	ids2 "github.com/avalido/mpc-controller/utils/ids"
	"github.com/pkg/errors"
)

var (
	cachedAllPubKeys   []types.MpcPublicKey = nil
	cachedAllAddresses *ids.ShortSet        = nil
)

func LoadGroup(db core.Store, groupID [32]byte) (*types.Group, error) {
	key := []byte(DbKeyPrefixGroup)
	key = append(key, groupID[:]...)
	groupBytes, err := db.Get(context.Background(), key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load group")
	}

	group := &types.Group{}
	err = group.Decode(groupBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode group: %v %v", key, groupBytes)
	}
	return group, nil
}

func LoadLatestPubKey(db core.Store) (*types.MpcPublicKey, error) {
	bytes, err := db.Get(context.Background(), []byte(DbKeyLatestPubKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load latest mpc public key")
	}

	if bytes == nil {
		return nil, errors.New("loaded empty mpc public key")
	}

	model := types.MpcPublicKey{}
	err = model.Decode(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode mpc public key")
	}

	return &model, nil
}

func SaveLatestPubKey(db core.Store, pubKey *types.MpcPublicKey) error {
	pubKeyBytes, err := pubKey.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode public key")
	}
	err = db.Set(context.Background(), []byte(DbKeyLatestPubKey), pubKeyBytes)
	if err != nil {
		return errors.Wrap(err, "failed to save latest public key")
	}
	return nil
}

func LoadAllAddresses(db core.Store) (*ids.ShortSet, error) {
	if cachedAllAddresses == nil {
		_, err := LoadAllPubKeys(db)
		if err != nil {
			return nil, err
		}
	}
	return cachedAllAddresses, nil
}

func LoadAllPubKeys(db core.Store) ([]types.MpcPublicKey, error) {
	if cachedAllPubKeys != nil {
		return cachedAllPubKeys, nil
	}
	bytes, err := db.Get(context.Background(), []byte(DbKeyAllPubKeys))
	if err != nil {
		return nil, nil
	}

	if bytes == nil {
		return nil, errors.New("loaded empty mpc public keys")
	}
	var pubKeys []types.MpcPublicKey
	err = json.Unmarshal(bytes, &pubKeys)
	if err != nil {
		return nil, errors.New("unable to decode mpc public keys")
	}
	err = refreshCache(pubKeys)
	if err != nil {
		return nil, err
	}
	return pubKeys, nil
}

func SaveAllPubKeys(db core.Store, pubKeys []types.MpcPublicKey) error {
	bytes, err := json.Marshal(pubKeys)
	if err != nil {
		return errors.Wrap(err, "failed to encode public keys")
	}
	err = db.Set(context.Background(), []byte(DbKeyAllPubKeys), bytes)
	if err != nil {
		return errors.Wrap(err, "failed to save public keys")
	}
	return refreshCache(pubKeys)
}

func AddPubKey(db core.Store, pubKey types.MpcPublicKey) error {
	pubKeys, err := LoadAllPubKeys(db)
	if err != nil {
		return err
	}
	pubKeys = append(pubKeys, pubKey)
	return SaveAllPubKeys(db, pubKeys)
}

func LoadGroupByLatestMpcPubKey(db core.Store) (*types.Group, error) {
	pubKey, err := LoadLatestPubKey(db)
	if err != nil {
		return nil, err
	}

	group, err := LoadGroup(db, pubKey.GroupId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load group")
	}

	return group, nil
}

func refreshCache(pubKeys []types.MpcPublicKey) error {
	cachedAllPubKeys = pubKeys
	err := refreshAddresses()
	if err != nil {
		return err
	}
	return nil
}

func refreshAddresses() error {
	shortSet := &ids.ShortSet{}
	for _, pubKey := range cachedAllPubKeys {
		genPK, err := crypto.NormalizePubKeyBytes(pubKey.GenPubKey)
		if err != nil {
			return err
		}
		id, err := ids2.ShortIDFromPubKeyBytes(genPK)
		if err != nil {
			return err
		}
		shortSet.Add(*id)
	}
	cachedAllAddresses = shortSet
	return nil
}
