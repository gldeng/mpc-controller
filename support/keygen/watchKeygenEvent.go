package keygen

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"time"
)

// Accept event: *events.ContractFiltererCreatedEvent
// Accept event: *events.GroupInfoStoredEvent

// Emit event: *contract.MpcManagerKeygenRequestAdded

type KeygenRequestAddedEventWatcher struct {
	ContractAddr common.Address
	Logger       logger.Logger
	Publisher    dispatcher.Publisher

	groupIdBytes [][32]byte
	filterer     bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerKeygenRequestAdded
	done chan struct{}
}

func (eh *KeygenRequestAddedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererCreatedEvent:
		eh.filterer = evt.Filterer
	case *events.GroupInfoStoredEvent:
		eh.groupIdBytes = append(eh.groupIdBytes, bytes.HexTo32Bytes(evt.Val.GroupIdHex))
	}
	if len(eh.groupIdBytes) > 0 {
		eh.doWatchKeygenRequestAdded(ctx)
	}
}

func (eh *KeygenRequestAddedEventWatcher) doWatchKeygenRequestAdded(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerKeygenRequestAdded)
	err := eh.subscribeKeygenRequestAdded(ctx, newSink, eh.groupIdBytes)
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.watchKeygenRequestAdded(ctx)
	}
}

func (eh *KeygenRequestAddedEventWatcher) subscribeKeygenRequestAdded(ctx context.Context, sink chan<- *contract.MpcManagerKeygenRequestAdded, groupIds [][32]byte) (err error) {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			return true, errors.WithStack(err)
		}

		newSub, err := filter.WatchKeygenRequestAdded(nil, sink, groupIds)
		if err != nil {
			return true, errors.WithStack(err)
		}

		eh.sub = newSub
		return false, nil
	})
	err = errors.Wrapf(err, "failed to subscribe KeygenRequestAdded event")
	return
}

func (eh *KeygenRequestAddedEventWatcher) watchKeygenRequestAdded(ctx context.Context) {
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
				eh.Logger.ErrorOnError(err, "Got an error during watching KeygenRequestAdded event")
			}
		}
	}()
}
