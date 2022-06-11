package queue

import "sync"

type ArrayQueue struct {
	maxLen int
	q      []interface{}
	mu     *sync.RWMutex
}

func NewArrayQueue(maxLen int) *ArrayQueue {
	if maxLen <= 0 {
		panic("Max length limit for a queue shouldn't less than zero")
	}
	queue := make([]interface{}, 0, 64)
	return &ArrayQueue{
		maxLen: maxLen,
		q:      queue,
		mu:     new(sync.RWMutex),
	}
}

func (q *ArrayQueue) Enqueue(e interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.q) >= q.maxLen {
		panic("The queue is full") // todo: deal with panic more elegantly
	}
	q.q = append(q.q, e)
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

func (q *ArrayQueue) Count() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.q)
}

func (q *ArrayQueue) Empty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.q) == 0
}

func (q *ArrayQueue) Full() bool {
	q.mu.RUnlock()
	defer q.mu.RUnlock()

	return len(q.q) == q.maxLen
}

func (q *ArrayQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.q = make([]interface{}, 0, 64)
}
