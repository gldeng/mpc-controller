package core

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/pkg/errors"
	"time"
)

type TxIndex interface {
	// GetTxByType retrieves the tx for the request identified by request hash and type
	GetTxByType(requestHash [32]byte, typ string) (ids.ID, error)
	// SetTxByType sets the txID for the request identified by request hash and type
	SetTxByType(requestHash [32]byte, typ string, txID ids.ID) error
	// PurgeOlderThan removes all records older than the given time
	PurgeOlderThan(time time.Time) error
	// DeleteTxs deletes all txs for the request, used to purge info not needed
	DeleteTxs(requestHash [32]byte) error
	// IsKnownTx tells whether the txID is known to the index
	IsKnownTx(txID ids.ID) bool
}

var (
	_ TxIndex = (*InMemoryTxIndex)(nil)
)

type IDRecord struct {
	ID   ids.ID
	Time time.Time
}

type InMemoryTxIndex struct {
	index      map[[32]byte]map[string]IDRecord
	knownTxIds map[ids.ID]struct{}
}

func NewInMemoryTxIndex() *InMemoryTxIndex {
	return &InMemoryTxIndex{index: map[[32]byte]map[string]IDRecord{}}
}

func (i *InMemoryTxIndex) GetTxByType(requestHash [32]byte, typ string) (ids.ID, error) {
	byType, ok := i.index[requestHash]
	if !ok {
		return ids.Empty, errors.New("no tx is found for the request hash")
	}
	rec, ok := byType[typ]
	if !ok {
		return ids.Empty, errors.New("no tx is found for the type")
	}
	return rec.ID, nil
}

func (i *InMemoryTxIndex) SetTxByType(requestHash [32]byte, typ string, txID ids.ID) error {
	_, ok := i.index[requestHash]
	if !ok {
		i.index[requestHash] = map[string]IDRecord{}
	}
	i.index[requestHash][typ] = IDRecord{
		ID:   txID,
		Time: time.Now(),
	}
	i.knownTxIds[txID] = struct{}{}
	return nil
}

func (i *InMemoryTxIndex) PurgeOlderThan(time time.Time) error {
	for reqHash, m := range i.index {
		purgeAll := true
		for _, record := range m {
			purgeAll = purgeAll && record.Time.Before(time)
		}
		if purgeAll {
			for _, record := range m {
				delete(i.knownTxIds, record.ID)
			}
			err := i.DeleteTxs(reqHash)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *InMemoryTxIndex) DeleteTxs(requestHash [32]byte) error {
	delete(i.index, requestHash)
	return nil
}

func (i *InMemoryTxIndex) IsKnownTx(txID ids.ID) bool {
	_, ok := i.knownTxIds[txID]
	return ok
}
