package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"sync"
)

type Watcher struct {
	Logger  logger.Logger
	Arg     Arg
	closeCh chan struct{}
	wg      sync.WaitGroup
}

type Arg struct {
	Logs    chan types.Log
	Sub     event.Subscription
	Unpack  Unpack
	Process Process
}

type Unpack func(log types.Log) (evt interface{}, err error)
type Process func(ctx context.Context, evt interface{}) error

func (w *Watcher) Watch(ctx context.Context) error {
	w.closeCh = make(chan struct{})
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-ctx.Done():
				w.Arg.Sub.Unsubscribe()
				return
			case <-w.closeCh:
				w.Arg.Sub.Unsubscribe()
				return
			case log := <-w.Arg.Logs:
				evt, err := w.Arg.Unpack(log)
				if err != nil {
					w.Logger.ErrorOnError(err, "Failed to unpack log", []logger.Field{{"log", log}}...)
					break
				}
				w.Logger.ErrorOnError(w.Arg.Process(ctx, evt), "Failed to process log", []logger.Field{{"log", log}}...)
			case err := <-w.Arg.Sub.Err():
				w.Logger.ErrorOnError(err, "Got an watching error")
			}
		}
	}()
	return nil
}

func (w *Watcher) Close() {
	close(w.closeCh)
	w.wg.Wait()
}
