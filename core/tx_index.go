package core

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/pkg/errors"
)

type TxIndex interface {
	// GetTxByType retrieves the tx for the request identified by request hash and type
	GetTxByType(requestHash [32]byte, typ string) (ids.ID, error)
	// SetTxByType sets the txID for the request identified by request hash and type
	SetTxByType(requestHash [32]byte, typ string, txID ids.ID) error
	// DeleteTxs deletes all txs for the request, used to purge info not needed
	DeleteTxs(requestHash [32]byte) error
	// IsKnownTx tells whether the txID is known to the index
	IsKnownTx(txID ids.ID) bool
}

var (
	_ TxIndex = (*InMemoryTxIndex)(nil)
)

type InMemoryTxIndex struct {
	index      map[[32]byte]map[string]ids.ID
	knownTxIds map[ids.ID]struct{}
}

func NewInMemoryTxIndex() *InMemoryTxIndex {
	return &InMemoryTxIndex{index: map[[32]byte]map[string]ids.ID{}}
}

func (i *InMemoryTxIndex) GetTxByType(requestHash [32]byte, typ string) (ids.ID, error) {
	byType, ok := i.index[requestHash]
	if !ok {
		return ids.Empty, errors.New("no tx is found for the request hash")
	}
	id, ok := byType[typ]
	if !ok {
		return ids.Empty, errors.New("no tx is found for the type")
	}
	return id, nil
}

func (i *InMemoryTxIndex) SetTxByType(requestHash [32]byte, typ string, txID ids.ID) error {
	_, ok := i.index[requestHash]
	if !ok {
		i.index[requestHash] = map[string]ids.ID{}
	}
	i.index[requestHash][typ] = txID
	i.knownTxIds[txID] = struct{}{}
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
