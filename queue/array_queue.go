package queue

import (
	ctlPk "github.com/avalido/mpc-controller"
)

var _ ctlPk.Queue

type queueArray []any

func NewQueueArray() *queueArray {
	queue := make(queueArray, 0, 64)
	return &queue
}

func (q *queueArray) Enqueue(e any) {
	*q = append(*q, e)
}

func (q *queueArray) Dequeue() any {
	if len(*q) == 0 {
		return nil
	}
	e := (*q)[0]
	*q = (*q)[1:]
	return e
}

func (q *queueArray) Peek() any {
	if len(*q) == 0 {
		return nil
	}
	return (*q)[len(*q)-1]
}

func (q *queueArray) Count() int {
	return len(*q)
}

func (q *queueArray) Empty() bool {
	return len(*q) == 0
}

func (q *queueArray) Clear() {
	*q = make(queueArray, 0, 64)
}
