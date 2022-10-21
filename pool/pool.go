package pool

import (
	"github.com/alitto/pond"
	"github.com/enriquebris/goconcurrentqueue"
)

var (
	_ TaskSubmitter = (*ExtendedWorkerPool)(nil)
)

type ExtendedWorkerPool struct {
	inner     *pond.WorkerPool
	resources *goconcurrentqueue.FIFO
}

func New(size int, makeResources ResourcesFactory) (*ExtendedWorkerPool, error) {
	inner := pond.New(size, 1024)
	resources := goconcurrentqueue.NewFIFO()
	for i := 0; i < size; i++ {
		resources.Enqueue(makeResources()) // TODO: Construct resources
	}
	return &ExtendedWorkerPool{
		inner:     inner,
		resources: resources,
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
		res, _ := e.resources.Dequeue()        // TODO: Handle error
		next, _ := task.Next(res.(*Resources)) // TODO: Handle error
		e.resources.Enqueue(res)
		if next != nil {
			for _, t := range next {
				e.Submit(t) // Task needs to continue with itself or succeeding tasks
			}
		}
	}
	e.inner.Submit(taskWrapper)
	return nil
}
