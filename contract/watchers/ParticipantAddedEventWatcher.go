package watchers

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

type WatchParticipantAddedFilter interface {
	WatchParticipantAdded(opts *bind.WatchOpts, sink chan<- *contract.MpcManagerParticipantAdded, publicKey [][]byte) (event.Subscription, error)
}

// Trigger event: *events.MpcControllerPubKeyConfiguredEvent
// Emit event: *contract.MpcManagerParticipantAdded

type ParticipantAddedEventWatcher struct {
	Logger logger.Logger
	Filter func() WatchParticipantAddedFilter
	Signer *bind.WatchOpts

	Publisher dispatcher.Publisher

	pubKeyBytes [][]byte

	sub  event.Subscription
	sink chan *contract.MpcManagerParticipantAdded
	done chan struct{}
}

func (o *ParticipantAddedEventWatcher) Do(evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*events.MpcControllerPubKeyConfiguredEvent); ok {
		o.pubKeyBytes = append(o.pubKeyBytes, bytes.HexToBytes(evt.PartiPubKeyHex))

		newSink := make(chan *contract.MpcManagerParticipantAdded)
		err := o.subscribeParticipantAdded(evtObj.Context, newSink, o.pubKeyBytes)
		if err == nil {
			o.sink = newSink
			if o.done != nil {
				close(o.done)
			}
			o.done = make(chan struct{})
			o.watchParticipantAdded(evtObj.Context)
		}
	}
}

func (o *ParticipantAddedEventWatcher) subscribeParticipantAdded(ctx context.Context, sink chan<- *contract.MpcManagerParticipantAdded, pubKeys [][]byte) error {
	if o.sub != nil {
		o.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(o.Logger, ctx, func() error {
		newSub, err := o.Filter().WatchParticipantAdded(o.Signer, sink, pubKeys)
		if err != nil {
			o.Logger.Error("Failed to watch ParticipantAdded event", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		o.sub = newSub
		return nil
	})

	return err
}

func (o *ParticipantAddedEventWatcher) watchParticipantAdded(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-o.done:
				return
			case evt := <-o.sink:
				_ = evt
				evtObj := dispatcher.NewRootEventObject("ParticipantAddedEventWatcher", evt, ctx)
				o.Publisher.Publish(ctx, evtObj)

			case err := <-o.sub.Err():
				o.Logger.ErrorOnError(err, "Got an error during watching ParticipantAdded event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
