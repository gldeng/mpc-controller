package wrappers

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
)

type WatchKeygenRequestAddedFilter interface {
	WatchKeygenRequestAdded(opts *bind.WatchOpts, sink chan<- *contract.MpcManagerKeygenRequestAdded, groupId [][32]byte) (event.Subscription, error)
}

type OnGroupInfoStoredEvtHandler struct {
	Logger logger.Logger
	Filter func() WatchKeygenRequestAddedFilter
	Signer *bind.WatchOpts

	Publisher dispatcher.Publisher

	groupIdBytes [][32]byte

	sub  event.Subscription
	sink chan *contract.MpcManagerKeygenRequestAdded
	done chan struct{}
}

func (o *OnGroupInfoStoredEvtHandler) Do(evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*events.GroupInfoStoredEvent); ok {
		o.groupIdBytes = append(o.groupIdBytes, bytes.HexTo32Bytes(evt.GroupIdHex))

		newSink := make(chan *contract.MpcManagerKeygenRequestAdded)
		err := o.subscribeKeygenRequestAdded(evtObj.Context, newSink, o.groupIdBytes)
		if err == nil {
			o.sink = newSink
			if o.done != nil {
				close(o.done)
			}
			o.done = make(chan struct{})
			o.watchKeygenRequestAdded(evtObj.Context)
		}
	}
}

func (o *OnGroupInfoStoredEvtHandler) subscribeKeygenRequestAdded(ctx context.Context, sink chan<- *contract.MpcManagerKeygenRequestAdded, groupId [][32]byte) error {
	if o.sub != nil {
		o.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(o.Logger, ctx, func() error {
		newSub, err := o.Filter().WatchKeygenRequestAdded(o.Signer, sink, groupId)
		if err != nil {
			o.Logger.Error("Failed to watch KeygenRequestAdded event", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		o.sub = newSub
		return nil
	})

	newSub, err := o.Filter().WatchKeygenRequestAdded(o.Signer, sink, groupId)
	o.sub = newSub

	return err
}

func (o *OnGroupInfoStoredEvtHandler) watchKeygenRequestAdded(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-o.done:
				return
			case evt := <-o.sink:
				_ = evt
				evtObj := dispatcher.NewRootEventObject("OnGroupInfoStoredEvtHandler", evt, ctx)
				o.Publisher.Publish(ctx, evtObj)

			case err := <-o.sub.Err():
				o.Logger.ErrorOnError(err, "Got an error during watching KeygenRequestAdded event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
