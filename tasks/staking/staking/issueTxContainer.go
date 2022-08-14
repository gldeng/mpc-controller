package staking

import (
	"sort"
	"sync"
)

type issueTxContainer struct {
	sync.Mutex
	issueTxs issueTxs
}

func (c *issueTxContainer) AddSort(tx *issueTx) {
	c.Lock()
	defer c.Unlock()
	c.issueTxs = append(c.issueTxs, tx)
	c.sort()
}

func (c *issueTxContainer) ContinuousTxs(nonce uint64) []*issueTx {
	c.Lock()
	defer c.Unlock()
	indices := c.continuousIndices(nonce)
	var txs []*issueTx
	for _, i := range indices {
		txs = append(txs, c.issueTxs[i])
	}
	return txs
}

func (c *issueTxContainer) TrimLeft(nonce uint64) {
	c.Lock()
	defer c.Unlock()
	if !c.containNonce(nonce) {
		return
	}
	i := c.nonceIndex(nonce)
	c.issueTxs = c.issueTxs[i+1:]
}

func (c *issueTxContainer) IsEmpty() bool {
	c.Lock()
	defer c.Unlock()
	return len(c.issueTxs) == 0
}

func (c *issueTxContainer) Nonces() []uint64 {
	c.Lock()
	defer c.Unlock()
	var nonces []uint64
	for _, t := range c.issueTxs {
		nonces = append(nonces, t.Nonce)
	}
	return nonces
}

func (c *issueTxContainer) sort() {
	sort.Sort(c.issueTxs)
}

func (c *issueTxContainer) continuousIndices(nonce uint64) []int {
	if !c.containNonce(nonce) {
		return nil
	}

	var nextNonce = nonce
	var indices []int
	for i := c.nonceIndex(nonce); i < len(c.issueTxs); i++ {
		if c.issueTxs[i].Nonce != nextNonce {
			break
		}

		indices = append(indices, i)
		nextNonce++
	}
	return indices
}

func (c *issueTxContainer) containNonce(nonce uint64) bool {
	for _, tx := range c.issueTxs {
		if tx.Nonce == nonce {
			return true
		}
	}
	return false
}

func (c *issueTxContainer) nonceIndex(nonce uint64) int {
	var index = -1
	for i, tx := range c.issueTxs {
		if tx.Nonce == nonce {
			index = i
			break
		}
	}
	return index
}

// --
type issueTxs []*issueTx

func (t issueTxs) Len() int {
	return len(t)
}

func (t issueTxs) Less(i, j int) bool {
	return t[i].Nonce < t[j].Nonce
}

func (t issueTxs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
