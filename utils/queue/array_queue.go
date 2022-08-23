package queue

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

var _ Queue = (*ArrayQueue)(nil)

const (
	defaultMinLen = 64
	defaultMaxLen = 1024
)

type ArrayQueue struct {
	maxLen int
	q      []interface{}
	mu     *sync.RWMutex
}

func NewArrayQueue(maxLen int) *ArrayQueue {
	maxLen_ := maxLen
	if maxLen < defaultMinLen {
		maxLen_ = defaultMinLen
	}
	if maxLen_ > defaultMaxLen {
		maxLen_ = defaultMaxLen
	}

	queue := make([]interface{}, 0, 64)
	return &ArrayQueue{
		maxLen: maxLen_,
		q:      queue,
		mu:     new(sync.RWMutex),
	}
}

func (q *ArrayQueue) Enqueue(e interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.q) >= q.maxLen {
		return errors.WithStack(&ErrTypQueueIsFull{ErrMsg: fmt.Sprintf("%v. maxLen:%v", ErrMsgQueueIsFull, q.maxLen)})
	}
	q.q = append(q.q, e)
	return nil
}

func (q *ArrayQueue) Dequeue() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.q) == 0 {
		return nil
	}
	e := q.q[0]
	q.q = q.q[1:]
	return e
}

func (q *ArrayQueue) Peek() interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.q) == 0 {
		return nil
	}
	return (q.q)[len(q.q)-1]
}

func (q *ArrayQueue) List() []interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var list []interface{}
	for _, e := range q.q {
		list = append(list, e)
	}

	return list
}

func (q *ArrayQueue) Count() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.q)
}

func (q *ArrayQueue) Capacity() int {
	return q.maxLen
}

func (q *ArrayQueue) Empty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.q) == 0
}

func (q *ArrayQueue) Full() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.q) == q.maxLen
}

func (q *ArrayQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.q = make([]interface{}, 0, 64)
}
