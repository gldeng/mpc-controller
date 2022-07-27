package noncer

import (
	"sync"
)

type Noncer struct {
	BaseReqID uint64
	BaseNonce uint64
	Lock      sync.Mutex
}

func (n *Noncer) GetNonce(reqID uint64) (nonce uint64) {
	n.Lock.Lock()
	nonce = n.BaseNonce + (reqID - n.BaseReqID)
	n.Lock.Unlock()
	return
}

func (n *Noncer) ResetBase(baseReqID, baseNonce uint64) {
	n.Lock.Lock()
	n.BaseReqID = baseReqID
	n.BaseNonce = baseNonce
	n.Lock.Unlock()
}
