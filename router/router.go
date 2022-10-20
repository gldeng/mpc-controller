package router

import (
	"context"
	"sync"
)

type Queue interface {
	DequeueOrWaitForNextElementContext(ctx context.Context) (interface{}, error)
}
type Publisher interface {
	Publish(ctx context.Context, evt interface{})
}

type Router struct {
	unsub            chan struct{}
	closeOnce        sync.Once
	onCloseCtx       context.Context
	onCloseCtxCancel func()
	incomingQueue    Queue
	publisher        Publisher
}

func NewRouter(incomingQueue Queue, publisher Publisher) (*Router, error) {
	onCloseCtx, cancel := context.WithCancel(context.Background())
	return &Router{
		unsub:            make(chan struct{}),
		closeOnce:        sync.Once{},
		onCloseCtx:       onCloseCtx,
		onCloseCtxCancel: cancel,
		incomingQueue:    incomingQueue,
		publisher:        publisher,
	}, nil
}

func (r *Router) Start() error {
	go func() {
		event, _ := r.incomingQueue.DequeueOrWaitForNextElementContext(r.onCloseCtx) // TODO: Handle error
		r.publisher.Publish(r.onCloseCtx, event)
	}()
	return nil
}

func (r *Router) Close() error {
	r.closeOnce.Do(func() {
		r.unsub <- struct{}{}
		r.onCloseCtxCancel()
	})
	return nil
}
