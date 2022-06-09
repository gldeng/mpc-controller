package queue

import "sync"

var defaultQueue *queueArray

var once = &sync.Once{}

func initDefaultQueue() {
	once.Do(func() {
		defaultQueue = NewQueueArray()
	})
}

func Enqueue(e any) {
	initDefaultQueue()
	defaultQueue.Enqueue(e)
}

func Dequeue() any {
	initDefaultQueue()
	return defaultQueue.Dequeue()
}

func Peek() any {
	initDefaultQueue()
	return defaultQueue.Peek()
}

func Count() int {
	initDefaultQueue()
	return defaultQueue.Count()
}

func Empty() bool {
	initDefaultQueue()
	return defaultQueue.Empty()
}

func Clear() {
	initDefaultQueue()
	defaultQueue.Clear()
}
