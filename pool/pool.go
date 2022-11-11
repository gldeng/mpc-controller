package pool

import (
	"github.com/alitto/pond"
	"github.com/avalido/mpc-controller/core"
	"github.com/enriquebris/goconcurrentqueue"
)

var (
	_ core.TaskSubmitter = (*ExtendedWorkerPool)(nil)
)

type ExtendedWorkerPool struct {
	sequentialWorker *pond.WorkerPool
	parallelPool     *pond.WorkerPool
	contexts         *goconcurrentqueue.FIFO
}

func NewExtendedWorkerPool(size int, makeContext core.TaskContextFactory) (*ExtendedWorkerPool, error) {
	sequentialWorker := pond.New(1, 1024)
	parallelPool := pond.New(size, 1024)
	contexts := goconcurrentqueue.NewFIFO()
	for i := 0; i < size+1; i++ {
		contexts.Enqueue(makeContext())
	}
	return &ExtendedWorkerPool{
		sequentialWorker: sequentialWorker,
		parallelPool:     parallelPool,
		contexts:         contexts,
	}, nil
}

func (e *ExtendedWorkerPool) Start() error {
	// Do nothing
	return nil
}

func (e *ExtendedWorkerPool) Close() error {
	e.parallelPool.StopAndWait()
	return nil
}

func (e *ExtendedWorkerPool) Submit(task core.Task) error {
	whichPool := e.parallelPool
	if task.IsSequential() {
		whichPool = e.sequentialWorker
	}
	taskWrapper := func() {
		ctx, _ := e.contexts.Dequeue() // TODO: Handle error
		taskCtx := ctx.(core.TaskContext)
		next, err := task.Next(taskCtx) // TODO: Handle error
		if err != nil {
			taskCtx.GetLogger().Debugf("got an error for task %v, error:%v", task.GetId(), err)
		}
		if task.FailedPermanently() {
			taskCtx.GetLogger().Debugf("task %v failed permanently, error:%v", task.GetId(), err)
		}
		if task.IsDone() {
			taskCtx.GetLogger().Debugf("task %v was done", task.GetId())
		}
		if !task.IsDone() && !task.FailedPermanently() {
			e.Submit(task)
		}
		e.contexts.Enqueue(ctx)
		if next != nil {
			for _, t := range next {
				e.Submit(t) // Task needs to continue with itself or succeeding tasks
			}
		}
	}
	whichPool.Submit(taskWrapper)
	return nil
}
