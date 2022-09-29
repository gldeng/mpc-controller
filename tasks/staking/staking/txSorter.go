package staking

import (
	"github.com/ethereum/go-ethereum/common"
	"sort"
	"sync"
)

type txSorter struct {
	sync.Mutex
	tasks pendingStakeTasks
}

func (p *txSorter) AddSort(st *pendingStakeTask) {
	p.Lock()
	defer p.Unlock()
	p.tasks = append(p.tasks, st)
	p.sort()
}

func (c *txSorter) ContinuousTxs(nonce uint64) []*pendingStakeTask {
	c.Lock()
	defer c.Unlock()
	indices := c.continuousIndices(nonce)
	var txs []*pendingStakeTask
	for _, i := range indices {
		txs = append(txs, c.tasks[i])
	}
	return txs
}

func (c *txSorter) TrimLeft(nonce uint64) {
	c.Lock()
	defer c.Unlock()
	if !c.containNonce(nonce) {
		return
	}
	i := c.nonceIndex(nonce)
	c.tasks = c.tasks[i+1:]
}

func (c *txSorter) IsEmpty() bool {
	c.Lock()
	defer c.Unlock()
	return len(c.tasks) == 0
}

func (c *txSorter) Nonces() []uint64 {
	c.Lock()
	defer c.Unlock()
	var nonces []uint64
	for _, t := range c.tasks {
		nonces = append(nonces, t.stakeTask.Nonce)
	}
	return nonces
}

func (c *txSorter) Address() common.Address {
	return c.tasks[0].stakeTask.CChainAddress // todo: take key-rotation into consideration
}

func (c *txSorter) sort() {
	sort.Sort(c.tasks)
}

func (c *txSorter) continuousIndices(nonce uint64) []int {
	if !c.containNonce(nonce) {
		return nil
	}

	var nextNonce = nonce
	var indices []int
	for i := c.nonceIndex(nonce); i < len(c.tasks); i++ {
		if c.tasks[i].stakeTask.Nonce != nextNonce {
			break
		}

		indices = append(indices, i)
		nextNonce++
	}
	return indices
}

func (c *txSorter) containNonce(nonce uint64) bool {
	for _, tx := range c.tasks {
		if tx.stakeTask.Nonce == nonce {
			return true
		}
	}
	return false
}

func (c *txSorter) nonceIndex(nonce uint64) int {
	var index = -1
	for i, tx := range c.tasks {
		if tx.stakeTask.Nonce == nonce {
			index = i
			break
		}
	}
	return index
}

// --
type pendingStakeTasks []*pendingStakeTask

func (p pendingStakeTasks) Len() int {
	return len(p)
}

func (p pendingStakeTasks) Less(i, j int) bool {
	return p[i].stakeTask.Nonce < p[j].stakeTask.Nonce
}

func (p pendingStakeTasks) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
