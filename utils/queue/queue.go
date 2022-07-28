package queue

type Queue interface {
	Enqueue(e interface{})
	Dequeue() interface{}

	Peek() interface{}
	List() []interface{}

	Count() int
	Capacity() int

	Empty() bool
	Full() bool

	Clear()
}
