package router

import (
	"context"
	"sync"
)

type Queue interface {
	DequeueOrWaitForNextElementContext(ctx context.Context) (interface{}, error)
}

type EventHandler = func(interface{})

type Router struct {
	unsub            chan struct{}
	closeOnce        sync.Once
	onCloseCtx       context.Context
	onCloseCtxCancel func()
	incomingQueue    Queue
	handlers         []EventHandler
}

func NewRouter(incomingQueue Queue) (*Router, error) {
	onCloseCtx, cancel := context.WithCancel(context.Background())
	return &Router{
		unsub:            make(chan struct{}),
		closeOnce:        sync.Once{},
		onCloseCtx:       onCloseCtx,
		onCloseCtxCancel: cancel,
		incomingQueue:    incomingQueue,
		handlers:         make([]EventHandler, 0),
	}, nil
}

func (r *Router) Start() error {
	go func() {
		event, _ := r.incomingQueue.DequeueOrWaitForNextElementContext(r.onCloseCtx) // TODO: Handle error
		for _, handler := range r.handlers {
			handler(event)
		}
	}()
	return nil
}

func (r *Router) AddHandler(handler EventHandler) {
	r.handlers = append(r.handlers, handler)
}

func (r *Router) Close() error {
	r.closeOnce.Do(func() {
		r.unsub <- struct{}{}
		r.onCloseCtxCancel()
	})
	return nil
}
