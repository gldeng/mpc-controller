package pool

import (
	"github.com/alitto/pond"
	"github.com/avalido/mpc-controller/core"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/pkg/errors"
	"sync/atomic"
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
		err := contexts.Enqueue(makeContext())
		if err != nil {
			return nil, errors.Wrap(err, "failed to enqueue task context")
		}
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
			taskCtx.GetLogger().Debugf("%v got error:%v", task.GetId(), err)
		}
		if task.FailedPermanently() {
			taskCtx.GetLogger().Debugf("%v failed permanently, error:%v", task.GetId(), err)
			taskCtx.GetLogger().Debugf("NonceConsumers after task %v: %v", task.GetId(), atomic.LoadInt32(&core.NonceConsumers))
		}
		if task.IsDone() {
			taskCtx.GetLogger().Debugf("%v done", task.GetId())
			taskCtx.GetLogger().Debugf("NonceConsumers after task %v: %v", task.GetId(), atomic.LoadInt32(&core.NonceConsumers))
		}
		if !task.IsDone() && !task.FailedPermanently() {
			err = e.Submit(task)
			if err != nil {
				taskCtx.GetLogger().Debugf("failed to submit task, error:%v", err)
			}
		}
		err = e.contexts.Enqueue(ctx)
		if err != nil {
			taskCtx.GetLogger().Debugf("failed to enqueue task context, error:%v", err)
		}
		if next != nil {
			for _, t := range next {
				err = e.Submit(t) // Task needs to continue with itself or succeeding tasks
				if err != nil {
					taskCtx.GetLogger().Debugf("failed to submit task, error:%v", err)
				}
			}
		}
	}
	whichPool.Submit(taskWrapper)
	return nil
}
