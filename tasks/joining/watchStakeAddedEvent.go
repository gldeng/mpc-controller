package joining

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
)

// Accept event: *events.ContractFiltererCreatedEvent
// Accept event:  *events.GeneratedPubKeyInfoStoredEvent

// Emit event:  *contract.MpcManagerStakeRequestAdded

type StakeRequestAddedEventWatcher struct {
	Logger logger.Logger

	ContractAddr common.Address

	Publisher dispatcher.Publisher

	pubKeyBytes [][]byte
	filterer    bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerStakeRequestAdded
	done chan struct{}
}

func (eh *StakeRequestAddedEventWatcher) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererCreatedEvent:
		eh.filterer = evt.Filterer
	case *events.GeneratedPubKeyInfoStoredEvent:
		dnmPubKeyBtes, err := crypto.DenormalizePubKeyFromHex(evt.Val.GenPubKeyHex)
		if err != nil {
			eh.Logger.Error("Failed to denormalized generated public key", []logger.Field{{"error", err}}...)
			break
		}

		eh.pubKeyBytes = append(eh.pubKeyBytes, dnmPubKeyBtes)
	}
	if len(eh.pubKeyBytes) > 0 {
		eh.doWatchStakeRequestAdded(evtObj.Context)
	}
}

func (eh *StakeRequestAddedEventWatcher) doWatchStakeRequestAdded(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerStakeRequestAdded)
	err := eh.subscribeStakeRequestAdded(ctx, newSink, eh.pubKeyBytes)
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.watchStakeRequestAdded(ctx)
	}
}

func (eh *StakeRequestAddedEventWatcher) subscribeStakeRequestAdded(ctx context.Context, sink chan<- *contract.MpcManagerStakeRequestAdded, pubKeys [][]byte) error {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			eh.Logger.Error("Failed to create MpcManagerFilterer", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		newSub, err := filter.WatchStakeRequestAdded(nil, sink, pubKeys)
		if err != nil {
			eh.Logger.Error("Failed to watch StakeRequestAdded event", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		eh.sub = newSub
		return nil
	})

	return err
}

func (eh *StakeRequestAddedEventWatcher) watchStakeRequestAdded(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-eh.done:
				return
			case evt := <-eh.sink:
				evtObj := dispatcher.NewRootEventObject("StakeRequestAddedEventWatcher", evt, ctx)
				eh.Publisher.Publish(ctx, evtObj)
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching StakeRequestAdded event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
