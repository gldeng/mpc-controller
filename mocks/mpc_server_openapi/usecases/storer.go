package usecases

import "sync"

var storer = NewStorer()

type Storer struct {
	keygenMap    map[string]*KeygenRequestModel
	keygenLocker *sync.RWMutex

	signMap    map[string]*SignRequestModel
	signLocker *sync.RWMutex
}

func NewStorer() *Storer {
	return &Storer{
		keygenMap:    make(map[string]*KeygenRequestModel),
		keygenLocker: new(sync.RWMutex),

		signMap:    make(map[string]*SignRequestModel),
		signLocker: new(sync.RWMutex),
	}
}

// Operations on keygen request

func (s *Storer) GetKeygenRequestModel(requestId string) *KeygenRequestModel {
	s.keygenLocker.RLock()
	s.keygenLocker.RUnlock()

	return s.keygenMap[requestId]
}

func (s *Storer) StoreKeygenRequestModel(m *KeygenRequestModel) {
	s.keygenLocker.Lock()
	s.keygenLocker.Unlock()

	s.keygenMap[m.input.RequestId] = m
}

// Operations on sign request

func (s *Storer) GetSignRequestModel(requestId string) *SignRequestModel {
	s.signLocker.RLock()
	s.signLocker.RUnlock()

	return s.signMap[requestId]
}

func (s *Storer) StoreSignRequestModel(m *SignRequestModel) {
	s.signLocker.Lock()
	s.signLocker.Unlock()

	s.signMap[m.input.RequestId] = m
}
