package participant

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
)

// Accept event: *events.ContractFiltererCreatedEvent

// Emit event: *contract.MpcManagerParticipantAdded

type ParticipantAddedEventWatcher struct {
	Logger logger.Logger

	MyPubKeyBytes []byte
	ContractAddr  common.Address

	Publisher dispatcher.Publisher

	filterer bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerParticipantAdded
	done chan struct{}
}

func (eh *ParticipantAddedEventWatcher) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererCreatedEvent:
		eh.filterer = evt.Filterer
		eh.doWatchParticipantAdded(evtObj.Context)
	}
}

func (eh *ParticipantAddedEventWatcher) doWatchParticipantAdded(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerParticipantAdded)
	err := eh.subscribeParticipantAdded(ctx, newSink, [][]byte{eh.MyPubKeyBytes})
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.watchParticipantAdded(ctx)
	}
}

func (eh *ParticipantAddedEventWatcher) subscribeParticipantAdded(ctx context.Context, sink chan<- *contract.MpcManagerParticipantAdded, pubKeys [][]byte) error {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			eh.Logger.Error("Failed to create MpcManagerFilterer", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		newSub, err := filter.WatchParticipantAdded(nil, sink, pubKeys)
		if err != nil {
			eh.Logger.Error("Failed to watch StakeRequestStarted event", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		eh.sub = newSub
		return nil
	})

	return err
}

func (eh *ParticipantAddedEventWatcher) watchParticipantAdded(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-eh.done:
				return
			case evt := <-eh.sink:
				evtObj := dispatcher.NewRootEventObject("ParticipantAddedEventWatcher", evt, ctx)
				eh.Publisher.Publish(ctx, evtObj)
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching ParticipantAdded event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
