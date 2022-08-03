package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"sync"
)

type Watcher struct {
	Logger     logger.Logger
	Subscriber Subscriber
	sub        event.Subscription
	closeCh    chan struct{}
	wg         sync.WaitGroup
}

func (w *Watcher) Watch(ctx context.Context) error {
	w.closeCh = make(chan struct{})

	logs, sub, err := w.Subscriber.Subscribe(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to watch")
	}
	w.sub = sub
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				return
			case <-w.closeCh:
				sub.Unsubscribe()
				return
			case log := <-logs:
				if err := w.Subscriber.Process(ctx, log); err != nil {
					w.Logger.ErrorOnError(err, "Failed to process log", []logger.Field{{"log", log}}...)
				}
			case err := <-sub.Err():
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
