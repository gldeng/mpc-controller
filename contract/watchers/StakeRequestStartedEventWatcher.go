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

type WatchStakeRequestStartedFilter interface {
	WatchStakeRequestStarted(opts *bind.WatchOpts, sink chan<- *contract.MpcManagerStakeRequestStarted, publicKey [][]byte) (event.Subscription, error)
}

// Trigger event: *events.GeneratedPubKeyInfoStoredEvent
// Emit event: *contract.MpcManagerStakeRequestStarted

type StakeRequestStartedEventWatcher struct {
	Logger logger.Logger
	Filter func() WatchStakeRequestStartedFilter
	Signer *bind.WatchOpts

	Publisher dispatcher.Publisher

	pubKeyBytes [][]byte

	sub  event.Subscription
	sink chan *contract.MpcManagerStakeRequestStarted
	done chan struct{}
}

func (o *StakeRequestStartedEventWatcher) Do(evtObj *dispatcher.EventObject) {
	if evt, ok := evtObj.Event.(*events.GeneratedPubKeyInfoStoredEvent); ok {
		o.pubKeyBytes = append(o.pubKeyBytes, bytes.HexToBytes(evt.PubKeyHex))

		newSink := make(chan *contract.MpcManagerStakeRequestStarted)
		err := o.subscribeStakeRequestStarted(evtObj.Context, newSink, o.pubKeyBytes)
		if err == nil {
			o.sink = newSink
			if o.done != nil {
				close(o.done)
			}
			o.done = make(chan struct{})
			o.watchStakeRequestStarted(evtObj.Context)
		}
	}
}

func (o *StakeRequestStartedEventWatcher) subscribeStakeRequestStarted(ctx context.Context, sink chan<- *contract.MpcManagerStakeRequestStarted, pubKeys [][]byte) error {
	if o.sub != nil {
		o.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(o.Logger, ctx, func() error {
		newSub, err := o.Filter().WatchStakeRequestStarted(o.Signer, sink, pubKeys)
		if err != nil {
			o.Logger.Error("Failed to watch StakeRequestStarted event", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		o.sub = newSub
		return nil
	})

	return err
}

func (o *StakeRequestStartedEventWatcher) watchStakeRequestStarted(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-o.done:
				return
			case evt := <-o.sink:
				_ = evt
				evtObj := dispatcher.NewRootEventObject("StakeRequestStartedEventWatcher", evt, ctx)
				o.Publisher.Publish(ctx, evtObj)

			case err := <-o.sub.Err():
				o.Logger.ErrorOnError(err, "Got an error during watching StakeRequestStarted event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
