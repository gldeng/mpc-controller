package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
)

// Subscribe event:

// Process event: *events.StakeRequestAdded

type StakeRequestAdded struct {
	Logger    logger.Logger
	PubKeys   [][]byte
	Filterer  *contract.MpcManagerFilterer
	Publisher dispatcher.Publisher
	sub       event.Subscription
	closeCh   chan struct{}
}

func (w *StakeRequestAdded) Watch(ctx context.Context) error {
	sink := make(chan *contract.MpcManagerParticipantAdded)
	sub, err := w.Filterer.WatchParticipantAdded(nil, sink, w.PubKeys)
	if err != nil {
		return errors.Wrapf(err, "failed to watch StakeRequestAdded")
	}
	w.sub = sub
	w.closeCh = make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				return
			case <-w.closeCh:
				sub.Unsubscribe()
				return
			case evt := <-sink:
				evtObj := dispatcher.NewEvtObj((*events.ParticipantAdded)(evt), nil)
				w.Publisher.Publish(ctx, evtObj)
			case err := <-sub.Err():
				w.Logger.ErrorOnError(err, "Got an error watching StakeRequestAdded")
			}
		}
	}()
	return nil
}

func (w *StakeRequestAdded) Close() {
	close(w.closeCh)
}
