package router

import (
	"context"
	"github.com/avalido/mpc-controller/core"
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
	eventHandlerContext core.EventHandlerContext
	submitter           core.TaskSubmitter
}

func NewRouter(logger logger.Logger, incomingQueue Queue, eventHandlerContext core.EventHandlerContext, submitter core.TaskSubmitter) (*Router, error) {
	onCloseCtx, cancel := context.WithCancel(context.Background())
	return &Router{
		unsub:               make(chan struct{}),
		closeOnce:           sync.Once{},
		onCloseCtx:          onCloseCtx,
		onCloseCtxCancel:    cancel,
		incomingQueue:       incomingQueue,
		handlers:            make([]EventHandler, 0),
		logEventHandlers:    make([]core.LogEventHandler, 0),
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
					r.logger.Error("dequeue error", []logger.Field{{"error", err}}...)
					prom.QueueOperationError.With(prometheus.Labels{"pkg": "router", "operation": "dequeue"}).Inc()
					time.Sleep(time.Second)
					continue
				}
				for _, handler := range r.handlers {
					handler(event)
				}
				if evt, ok := event.(types.Log); ok {
					for _, handler := range r.logEventHandlers {
						tasks, _ := handler.Handle(r.eventHandlerContext, evt) // TODO: Handle error
						for _, task := range tasks {
							err := r.submitter.Submit(task)
							_ = err // TODO: handle error
						}
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

func (r *Router) Close() error {
	r.closeOnce.Do(func() {
		r.unsub <- struct{}{}
		r.onCloseCtxCancel()
	})
	return nil
}
