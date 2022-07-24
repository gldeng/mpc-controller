package staking

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
	"time"
)

// Accept event: *events.ContractFiltererCreatedEvent
// Accept event: *events.GeneratedPubKeyInfoStoredEvent

// Emit event: *contract.MpcManagerStakeRequestStarted

type StakeRequestStartedEventWatcher struct {
	Logger logger.Logger

	ContractAddr common.Address

	Publisher dispatcher.Publisher

	pubKeyBytes [][]byte
	filterer    bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerStakeRequestStarted
	done chan struct{}
}

func (eh *StakeRequestStartedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererCreatedEvent:
		eh.filterer = evt.Filterer
	case *events.GeneratedPubKeyInfoStoredEvent:
		dnmPubKeyBtes, err := crypto.DenormalizePubKeyFromHex(evt.Val.CompressedGenPubKeyHex)
		if err != nil {
			eh.Logger.Error("Failed to denormalized generated public key", []logger.Field{{"error", err}}...)
			break
		}

		eh.pubKeyBytes = append(eh.pubKeyBytes, dnmPubKeyBtes)
	}
	if len(eh.pubKeyBytes) > 0 {
		eh.doWatchStakeRequestStarted(evtObj.Context)
	}
}

func (eh *StakeRequestStartedEventWatcher) doWatchStakeRequestStarted(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerStakeRequestStarted, 1024) // todo: config capacity?
	err := eh.subscribeStakeRequestStarted(ctx, newSink, eh.pubKeyBytes)
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.watchStakeRequestStarted(ctx)
	}
}

func (eh *StakeRequestStartedEventWatcher) subscribeStakeRequestStarted(ctx context.Context, sink chan<- *contract.MpcManagerStakeRequestStarted, pubKeys [][]byte) (err error) {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}

	err = backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			return true, errors.WithStack(err)
		}

		newSub, err := filter.WatchStakeRequestStarted(nil, sink, pubKeys)
		if err != nil {
			return true, errors.WithStack(err)
		}
		eh.sub = newSub
		return false, nil
	})
	err = errors.Wrapf(err, "failed to subscribe StakeRequestStarted event")
	return
}

func (eh *StakeRequestStartedEventWatcher) watchStakeRequestStarted(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-eh.done:
				return
			case evt := <-eh.sink: // todo: take measures for event task tracking and restoring
				evtObj := dispatcher.NewRootEventObject("StakeRequestStartedEventWatcher", evt, ctx)
				eh.Publisher.Publish(ctx, evtObj)
				eh.Logger.Debug("Stake request started", []logger.Field{
					{"StakeRequestStartedEvent", evt}}...)
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching StakeRequestStarted event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
