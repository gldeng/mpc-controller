package noncer

import (
	"sync"
)

type Noncer interface {
	GetNonce(reqID uint64) (nonce uint64)
	ResetBase(baseReqID, baseNonce uint64) bool
}

type noncer struct {
	baseReqID uint64
	baseNonce uint64
	gap       int64
	lock      *sync.Mutex
}

func New(baseReqID, baseNonce uint64) Noncer {
	return &noncer{
		baseReqID: baseReqID,
		baseNonce: baseNonce,
		gap:       int64(baseReqID - baseNonce),
		lock:      new(sync.Mutex),
	}
}

func (n *noncer) GetNonce(reqID uint64) (nonce uint64) {
	n.lock.Lock()
	defer n.lock.Unlock()
	nonceInt64 := int64(reqID) - n.gap
	return uint64(nonceInt64)
}

func (n *noncer) ResetBase(baseReqID, baseNonce uint64) bool {
	n.lock.Lock()
	defer n.lock.Unlock()
	newGap := int64(baseReqID - baseNonce)
	if newGap == n.gap {
		return false
	}
	n.baseReqID = baseReqID
	n.baseNonce = baseNonce
	n.gap = newGap
	return true
}
