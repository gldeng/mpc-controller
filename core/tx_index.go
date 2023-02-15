package core

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type TxIndexReader interface {
	// GetTxByType retrieves the tx for the request identified by request hash and type
	GetTxByType(requestHash [32]byte, typ string) (ids.ID, error)
}

type TxIndex interface {
	TxIndexReader
	// SetTxByType sets the txID for the request identified by request hash and type
	SetTxByType(requestHash [32]byte, typ string, txID ids.ID) error
	// PurgeOlderThan removes all records older than the given time
	PurgeOlderThan(time time.Time) error
	// IsKnownTx tells whether the txID is known to the index
	IsKnownTx(txID ids.ID) bool
}

var (
	_ TxIndex = (*InMemoryTxIndex)(nil)
)

type IDRecord struct {
	ID   ids.ID    `json:"id"`
	Time time.Time `json:"time"`
}

type InMemoryTxIndex struct {
	index      map[string]map[string]IDRecord
	knownTxIds map[string]struct{}
	mutex      sync.Mutex
}

func NewInMemoryTxIndex() *InMemoryTxIndex {
	return &InMemoryTxIndex{
		index:      map[string]map[string]IDRecord{},
		knownTxIds: map[string]struct{}{},
		mutex:      sync.Mutex{},
	}
}

func (i *InMemoryTxIndex) GetTxByType(requestHash [32]byte, typ string) (ids.ID, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	byType, ok := i.index[bytes.Bytes32ToHex(requestHash)]
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
	reqHash := bytes.Bytes32ToHex(requestHash)
	_, ok := i.index[reqHash]
	if !ok {
		i.index[reqHash] = map[string]IDRecord{}
	}
	i.index[reqHash][typ] = IDRecord{
		ID:   txID,
		Time: time.Now(),
	}
	i.knownTxIds[txID.String()] = struct{}{}
	return nil
}

func (i *InMemoryTxIndex) PurgeOlderThan(time time.Time) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	for reqHash, m := range i.index {
		purgeAll := true
		for _, record := range m {
			purgeAll = purgeAll && record.Time.Before(time)
		}
		if purgeAll {
			for _, record := range m {
				delete(i.knownTxIds, record.ID.String())
			}
			err := i.deleteTxs(reqHash)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *InMemoryTxIndex) deleteTxs(requestHash string) error {
	delete(i.index, requestHash)
	return nil
}

func (i *InMemoryTxIndex) IsKnownTx(txID ids.ID) bool {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	_, ok := i.knownTxIds[txID.String()]
	return ok
}
