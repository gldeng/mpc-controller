package noncer

import (
	"sync"
)

type Noncer interface {
	GetNonce(reqID uint64) (nonce uint64)
	ResetBase(baseReqID, baseNonce uint64)
}

type noncer struct {
	baseReqID uint64
	baseNonce uint64
	lock      *sync.Mutex
}

func New(baseReqID, baseNonce uint64) Noncer {
	return &noncer{
		baseReqID: baseReqID,
		baseNonce: baseNonce,
		lock:      new(sync.Mutex),
	}
}

func (n *noncer) GetNonce(reqID uint64) (nonce uint64) {
	n.lock.Lock()
	nonce = n.baseNonce + (reqID - n.baseReqID)
	n.lock.Unlock()
	return
}

func (n *noncer) ResetBase(baseReqID, baseNonce uint64) {
	n.lock.Lock()
	n.baseReqID = baseReqID
	n.baseNonce = baseNonce
	n.lock.Unlock()
}
