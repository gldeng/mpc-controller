package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"sync"
)

type Watcher struct {
	Logger    logger.Logger
	Subscribe Subscribe
	Publish   Publish
	Filterer  interface{}
	Publisher dispatcher.Publisher
	sub       event.Subscription
	closeCh   chan struct{}
	wg        sync.WaitGroup
}

func (w *Watcher) Watch(ctx context.Context) error {
	w.closeCh = make(chan struct{})

	sink, sub, err := w.Subscribe(w.Logger, ctx, w.closeCh, w.Filterer)
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
			case evt := <-sink:
				w.Publish(w.Logger, ctx, w.Publisher, evt)
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
