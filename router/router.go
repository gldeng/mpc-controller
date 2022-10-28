package router

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/ethereum/go-ethereum/core/types"
	"sync"
)

type Queue interface {
	DequeueOrWaitForNextElementContext(ctx context.Context) (interface{}, error)
}

type EventHandler = func(interface{})

type Router struct {
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

func NewRouter(incomingQueue Queue, eventHandlerContext core.EventHandlerContext, submitter core.TaskSubmitter) (*Router, error) {
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
		event, _ := r.incomingQueue.DequeueOrWaitForNextElementContext(r.onCloseCtx) // TODO: Handle error
		for _, handler := range r.handlers {
			handler(event)
		}
		if evt, ok := event.(types.Log); ok {
			for _, handler := range r.logEventHandlers {
				tasks, _ := handler.Handle(r.eventHandlerContext, evt) // TODO: Handle error
				for _, task := range tasks {
					r.submitter.Submit(task)
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
