package pool

import (
	"github.com/alitto/pond"
	"github.com/enriquebris/goconcurrentqueue"
)

var (
	_ TaskSubmitter = (*ExtendedWorkerPool)(nil)
)

type ExtendedWorkerPool struct {
	inner    *pond.WorkerPool
	contexts *goconcurrentqueue.FIFO
}

func New(size int, makeContext TaskContextFactory) (*ExtendedWorkerPool, error) {
	inner := pond.New(size, 1024)
	contexts := goconcurrentqueue.NewFIFO()
	for i := 0; i < size; i++ {
		contexts.Enqueue(makeContext())
	}
	return &ExtendedWorkerPool{
		inner:    inner,
		contexts: contexts,
	}, nil
}

func (e *ExtendedWorkerPool) Start() error {
	// Do nothing
	return nil
}

func (e *ExtendedWorkerPool) Close() error {
	e.inner.StopAndWait()
	return nil
}

func (e *ExtendedWorkerPool) Submit(task Task) error {
	taskWrapper := func() {
		ctx, _ := e.contexts.Dequeue()          // TODO: Handle error
		next, _ := task.Next(ctx.(TaskContext)) // TODO: Handle error
		e.contexts.Enqueue(ctx)
		if next != nil {
			for _, t := range next {
				e.Submit(t) // Task needs to continue with itself or succeeding tasks
			}
		}
	}
	e.inner.Submit(taskWrapper)
	return nil
}
