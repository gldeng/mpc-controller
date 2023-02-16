package router

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	types2 "github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

type Queue interface {
	Dequeue() (interface{}, error)
}

type EventHandler = func(interface{})

type Router struct {
	logger              logger.Logger
	unsub               chan struct{}
	closeOnce           sync.Once
	onCloseCtx          context.Context
	onCloseCtxCancel    func()
	incomingQueue       Queue
	handlers            []EventHandler
	logEventHandlers    []core.LogEventHandler
	pChainEventHandlers []core.PChainEventHandler
	eventHandlerContext core.EventHandlerContext
	submitter           core.TaskSubmitter
}

func NewRouter(logger logger.Logger, incomingQueue Queue, eventHandlerContext core.EventHandlerContext, submitter core.TaskSubmitter) (*Router, error) {
	onCloseCtx, cancel := context.WithCancel(context.Background())
	return &Router{
		logger:              logger,
		unsub:               make(chan struct{}),
		closeOnce:           sync.Once{},
		onCloseCtx:          onCloseCtx,
		onCloseCtxCancel:    cancel,
		incomingQueue:       incomingQueue,
		handlers:            make([]EventHandler, 0),
		logEventHandlers:    make([]core.LogEventHandler, 0),
		pChainEventHandlers: make([]core.PChainEventHandler, 0),
		eventHandlerContext: eventHandlerContext,
		submitter:           submitter,
	}, nil
}

func (r *Router) Start() error {
	go func() {
		for {
			select {
			case <-r.onCloseCtx.Done():
				return
			default:
				event, err := r.incomingQueue.Dequeue()
				if err != nil {
					// Here we take the following assumptions:
					// 1. The incomingQueue is a FIFO queue from https://github.com/enriquebris/goconcurrentqueue#fifo;
					// 2. The given FIFO queue will never get locked;
					// 3. It's a normal condition for the queue to stay at empty status;
					// 4. The error string must be empty or "The queue is locked" or "empty queue" (according to https://github.com/enriquebris/goconcurrentqueue/blob/master/fifo_queue.go#L82).
					// With the above assumptions in mind we don't need to take extra measures other than sleep for some time.
					// It's also not meaningful either to give error log or to update error metric,

					//r.logger.Error("dequeue error", []logger.Field{{"error", err}}...)
					//prom.QueueOperationError.With(prometheus.Labels{"pkg": "router", "operation": "dequeue"}).Inc()
					time.Sleep(time.Second)
					continue
				}
				prom.QueueOperation.With(prometheus.Labels{"pkg": "router", "operation": "dequeue"}).Inc()
				for _, handler := range r.handlers {
					handler(event)
				}
				if evt, ok := event.(types.Log); ok {
					for _, handler := range r.logEventHandlers {
						tasks, err := handler.Handle(r.eventHandlerContext, evt)
						if err != nil {
							r.logger.Error("handle eth log error", []logger.Field{{"error", err}}...)
							prom.HandleEthLogErr.Inc()
						}
						r.submitTasks(tasks)
					}
				}
				if evt, ok := event.(types2.UtxoBucket); ok {
					r.logger.Debugf("utxoBucket Event %v", evt)
					for _, handler := range r.pChainEventHandlers {
						tasks, err := handler.Handle(r.eventHandlerContext, evt)
						if err != nil {
							r.logger.Error("handle eth log error", []logger.Field{{"error", err}}...)
							prom.HandleEthLogErr.Inc()
						}
						r.submitTasks(tasks)
					}
				}
			}
		}
	}()
	return nil
}

func (r *Router) AddHandler(handler EventHandler) {
	r.handlers = append(r.handlers, handler)
}

func (r *Router) AddLogEventHandler(handler core.LogEventHandler) {
	r.logEventHandlers = append(r.logEventHandlers, handler)
}

func (r *Router) AddPChainEventHandler(handler core.PChainEventHandler) {
	r.pChainEventHandlers = append(r.pChainEventHandlers, handler)
}

func (r *Router) Close() error {
	r.closeOnce.Do(func() {
		r.unsub <- struct{}{}
		r.onCloseCtxCancel()
	})
	return nil
}

func (r *Router) submitTasks(tasks []core.Task) {
	for _, task := range tasks {
		err := r.submitter.Submit(task)
		if err != nil {
			prom.TaskSubmissionErr.Inc()
			r.logger.Error("task submission error", []logger.Field{
				{"id", task.GetId()},
				{"isDone", task.IsDone()},
				{"failedPermanently", task.FailedPermanently()},
				{"isSequential", task.IsSequential()},
				{"error", err}}...)
		}
	}
}
