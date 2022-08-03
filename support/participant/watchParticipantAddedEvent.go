package participant

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"time"
)

// Subscribe event: *events.ContractFiltererCreated

// Publish event: *contract.MpcManagerParticipantAdded

type ParticipantAddedEventWatcher struct {
	ContractAddr  common.Address
	Logger        logger.Logger
	MyPubKeyBytes []byte
	Publisher     dispatcher.Publisher

	filterer bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerParticipantAdded
	done chan struct{}
}

func (eh *ParticipantAddedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererCreated:
		eh.filterer = evt.Filterer
		eh.doWatchParticipantAdded(ctx)
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

func (eh *ParticipantAddedEventWatcher) subscribeParticipantAdded(ctx context.Context, sink chan<- *contract.MpcManagerParticipantAdded, pubKeys [][]byte) (err error) {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}
	err = backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			return true, errors.WithStack(err)
		}

		newSub, err := filter.WatchParticipantAdded(nil, sink, pubKeys)
		if err != nil {
			return true, errors.WithStack(err)
		}

		eh.sub = newSub
		return false, nil
	})
	err = errors.Wrapf(err, "failed to subscribe StakeRequestStarted event")
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
				evtObj := dispatcher.NewEvtObj(evt, nil)
				eh.Publisher.Publish(ctx, evtObj)
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching ParticipantAdded event")
			}
		}
	}()
}
