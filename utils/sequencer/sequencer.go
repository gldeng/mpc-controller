package sequencer

import (
	"sort"
	"sync"
)

type Sequencer interface {
	AddThenSort(o Obj)
	Objs() Objs
	ObjsFromNonce(nonce uint64) Objs
	TrimLeft(nonce uint64)
	IsEmpty() bool
}

type Obj interface {
	Nonce() uint64
}

type AscendingSequencer struct {
	sync.Mutex
	objs Objs
}

func (s *AscendingSequencer) AddThenSort(o Obj) {
	s.Lock()
	defer s.Unlock()
	s.objs = append(s.objs, o)
	s.sort()
}

func (s *AscendingSequencer) Objs() Objs {
	s.Lock()
	defer s.Unlock()
	return s.objs
}

func (s *AscendingSequencer) ObjsFromNonce(nonce uint64) Objs {
	s.Lock()
	defer s.Unlock()
	indices := s.continuousIndices(nonce)
	var objs []Obj
	for _, i := range indices {
		objs = append(objs, s.objs[i])
	}
	return objs
}

func (s *AscendingSequencer) TrimLeft(nonce uint64) {
	s.Lock()
	defer s.Unlock()
	if !s.containObj(nonce) {
		return
	}
	i := s.objIndex(nonce)
	s.objs = s.objs[i+1:]
}

func (s *AscendingSequencer) IsEmpty() bool {
	s.Lock()
	defer s.Unlock()
	return len(s.objs) == 0
}

func (s *AscendingSequencer) sort() {
	sort.Sort(s.objs)
}

func (s *AscendingSequencer) continuousIndices(nonce uint64) []int {
	if !s.containObj(nonce) {
		return nil
	}

	var nextNonce = nonce
	var indices []int
	for i := s.objIndex(nonce); i < len(s.objs); i++ {
		if s.objs[i].Nonce() != nextNonce {
			break
		}

		indices = append(indices, i)
		nextNonce++
	}
	return indices
}

func (s *AscendingSequencer) containObj(nonce uint64) bool {
	for _, obj := range s.objs {
		if obj.Nonce() == nonce {
			return true
		}
	}
	return false
}

func (s *AscendingSequencer) objIndex(nonce uint64) int {
	var index = -1
	for i, obj := range s.objs {
		if obj.Nonce() == nonce {
			index = i
			break
		}
	}
	return index
}

// --

type Objs []Obj

func (o Objs) Len() int {
	return len(o)
}

func (o Objs) Less(i, j int) bool {
	return o[i].Nonce() < o[j].Nonce()
}

func (o Objs) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
