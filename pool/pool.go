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
	logger           logger.Logger
}

func NewExtendedWorkerPool(size int, makeContext core.TaskContextFactory, logger logger.Logger) (*ExtendedWorkerPool, error) {
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
		logger:           logger,
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
	executeTaskAndMaybeSubmitContinuation := func() {
		continuation := e.getContextAndExecuteTask(task)
		for _, t := range continuation {
			err := e.Submit(t) // Note: This has to be after other task logics complete.
			if err != nil {
				e.logger.Debug("failed to submit task", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
			}
		}
	}
	whichPool.Submit(executeTaskAndMaybeSubmitContinuation)
	return nil
}

func (e *ExtendedWorkerPool) getContextAndExecuteTask(task core.Task) []core.Task {
	ctx, err := e.contexts.Dequeue()
	if err != nil {
		prom.QueueOperationError.With(prometheus.Labels{"pkg": "pool", "operation": "dequeue"}).Inc()
		panic(fmt.Sprintf("failed to submit task %v, dequeue error: %v", task.GetId(), err))
	}
	prom.QueueOperation.With(prometheus.Labels{"pkg": "pool", "operation": "dequeue"}).Inc()
	defer func() {
		err = e.contexts.Enqueue(ctx)
		if err != nil {
			prom.QueueOperationError.With(prometheus.Labels{"pkg": "pool", "operation": "enqueue"}).Inc()
			e.logger.Fatal("failed to enqueue task context, enqueue error", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
		}
		prom.QueueOperation.With(prometheus.Labels{"pkg": "pool", "operation": "enqueue"}).Inc()
	}()
	return executeTaskWithContext(task, ctx.(core.TaskContext))
}

func executeTaskWithContext(task core.Task, ctx core.TaskContext) (continuation []core.Task) {
	next, err := task.Next(ctx) // TODO: Handle error
	if err != nil {
		ctx.GetLogger().Debug("task got error", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
	}
	if task.FailedPermanently() {
		ctx.GetLogger().Debug("task failed permanently", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
	}
	if task.IsDone() {
		ctx.GetLogger().Debug("task done", []logger.Field{{"task", task.GetId()}}...)
	}
	continuation = append(continuation, next...)
	if !task.IsDone() && !task.FailedPermanently() {
		continuation = append(continuation, task)
		if err != nil {
			ctx.GetLogger().Debug("failed to submit task", []logger.Field{{"task", task.GetId()}, {"error", err}}...)
		}
	}
	return
}
