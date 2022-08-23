package usecases

import "sync"

var storer *Storer

type Storer struct {
	keygenMap    map[string]*KeygenRequestModel
	keygenLocker *sync.Mutex

	signMap    map[string]*SignRequestModel
	signLocker *sync.Mutex
}

func NewStorer() *Storer {
	return &Storer{
		keygenMap:    make(map[string]*KeygenRequestModel),
		keygenLocker: new(sync.Mutex),

		signMap:    make(map[string]*SignRequestModel),
		signLocker: new(sync.Mutex),
	}
}

// Operations on keygen request

func (s *Storer) GetKeygenRequestModel(requestId string) *KeygenRequestModel {
	s.keygenLocker.Lock()
	defer s.keygenLocker.Unlock()

	return s.keygenMap[requestId]
}

func (s *Storer) StoreKeygenRequestModel(m *KeygenRequestModel) {
	s.keygenLocker.Lock()
	defer s.keygenLocker.Unlock()

	s.keygenMap[m.input.RequestId] = m
}

// Operations on sign request

func (s *Storer) GetSignRequestModel(requestId string) *SignRequestModel {
	s.signLocker.Lock()
	defer s.signLocker.Unlock()

	return s.signMap[requestId]
}

func (s *Storer) StoreSignRequestModel(m *SignRequestModel) {
	s.signLocker.Lock()
	defer s.signLocker.Unlock()

	s.signMap[m.input.RequestId] = m
}

func init() {
	if storer == nil {
		storer = NewStorer()
	}
}
