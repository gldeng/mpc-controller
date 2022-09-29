package staking

import (
	"github.com/ethereum/go-ethereum/common"
	"sort"
	"sync"
)

type exportTxSorter struct {
	sync.Mutex
	tasks pendingStakeTasks
}

func (p *exportTxSorter) AddSort(st *pendingStakeTask) {
	p.Lock()
	defer p.Unlock()
	p.tasks = append(p.tasks, st)
	p.sort()
}

func (c *exportTxSorter) ContinuousTxs(nonce uint64) []*pendingStakeTask {
	c.Lock()
	defer c.Unlock()
	indices := c.continuousIndices(nonce)
	var txs []*pendingStakeTask
	for _, i := range indices {
		txs = append(txs, c.tasks[i])
	}
	return txs
}

func (c *exportTxSorter) TrimLeft(nonce uint64) {
	c.Lock()
	defer c.Unlock()
	if !c.containNonce(nonce) {
		return
	}
	i := c.nonceIndex(nonce)
	c.tasks = c.tasks[i+1:]
}

func (c *exportTxSorter) IsEmpty() bool {
	c.Lock()
	defer c.Unlock()
	return len(c.tasks) == 0
}

func (c *exportTxSorter) Nonces() []uint64 {
	c.Lock()
	defer c.Unlock()
	var nonces []uint64
	for _, t := range c.tasks {
		nonces = append(nonces, t.stakeTask.Nonce)
	}
	return nonces
}

func (c *exportTxSorter) Address() common.Address {
	return c.tasks[0].stakeTask.CChainAddress // todo: take key-rotation into consideration
}

func (c *exportTxSorter) sort() {
	sort.Sort(c.tasks)
}

func (c *exportTxSorter) continuousIndices(nonce uint64) []int {
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

func (c *exportTxSorter) containNonce(nonce uint64) bool {
	for _, tx := range c.tasks {
		if tx.stakeTask.Nonce == nonce {
			return true
		}
	}
	return false
}

func (c *exportTxSorter) nonceIndex(nonce uint64) int {
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
