package pool

import (
	"fmt"
	"github.com/alitto/pond"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	poolTypeSequential = "sequential_"
	poolTypeParallel   = "parallel_"
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
			prom.QueueOperationError.With(prometheus.Labels{"pkg": "pool", "operation": "enqueue"}).Inc()
			return nil, errors.Wrap(err, "failed to enqueue task context, enqueue error")
		}
	}
	prom.ConfigWorkPoolAndTaskMetrics(poolTypeSequential, sequentialWorker)
	prom.ConfigWorkPoolAndTaskMetrics(poolTypeParallel, parallelPool)
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
		ctx, err := e.contexts.Dequeue()
		if err != nil {
			prom.QueueOperationError.With(prometheus.Labels{"pkg": "pool", "operation": "dequeue"}).Inc()
			panic(fmt.Sprintf("failed to submit task %v, dequeue error: %v", task.GetId(), err))
		}
		prom.QueueOperation.With(prometheus.Labels{"pkg": "pool", "operation": "dequeue"}).Inc()
		taskCtx := ctx.(core.TaskContext)
		next, err := task.Next(taskCtx) // TODO: Handle error
		if err != nil {
			taskCtx.GetLogger().Debug("task got error", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
		}
		if task.FailedPermanently() {
			taskCtx.GetLogger().Debug("task failed permanently", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
		}
		if task.IsDone() {
			taskCtx.GetLogger().Debug("task done", []logger.Field{{"task", task.GetId()}}...)
		}
		if !task.IsDone() && !task.FailedPermanently() {
			err = e.Submit(task)
			if err != nil {
				taskCtx.GetLogger().Debug("failed to submit task", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
			}
		}
		err = e.contexts.Enqueue(ctx)
		if err != nil {
			prom.QueueOperationError.With(prometheus.Labels{"pkg": "pool", "operation": "enqueue"}).Inc()
			taskCtx.GetLogger().Fatal("failed to enqueue task context, enqueue error", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
		}
		prom.QueueOperation.With(prometheus.Labels{"pkg": "pool", "operation": "enqueue"}).Inc()
		if next != nil {
			for _, t := range next {
				err = e.Submit(t) // Task needs to continue with itself or succeeding tasks
				if err != nil {
					taskCtx.GetLogger().Debug("failed to submit task", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
				}
			}
		}
	}
	whichPool.Submit(taskWrapper)
	return nil
}
