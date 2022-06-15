package staking

import (
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// Accept event: *contract.MpcManagerStakeRequestStarted

// Emit event: ?

type StakeRequestStartedEventHandler struct {
	Logger logger.Logger

	ContractAddr common.Address

	Publisher dispatcher.Publisher

	pubKeyBytes [][]byte
	filterer    bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerStakeRequestStarted
	done chan struct{}
}

func (eh *StakeRequestStartedEventHandler) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererReconnectedEvent:
		eh.filterer = evt.Filterer
		eh.doWatchStakeRequestStarted(evtObj.Context)
	case *events.GeneratedPubKeyInfoStoredEvent:
		eh.pubKeyBytes = append(eh.pubKeyBytes, bytes.HexToBytes(evt.PubKeyHex))
		eh.doWatchStakeRequestStarted(evtObj.Context)
	}
}

//func (eh *StakeRequestStartedEventHandler) doWatchStakeRequestStarted(ctx context.Context) {
//	newSink := make(chan *contract.MpcManagerStakeRequestStarted)
//	err := eh.subscribeStakeRequestStarted(ctx, newSink, eh.pubKeyBytes)
//	if err == nil {
//		eh.sink = newSink
//		if eh.done != nil {
//			close(eh.done)
//		}
//		eh.done = make(chan struct{})
//		eh.watchStakeRequestStarted(ctx)
//	}
//}
//
//func (eh *StakeRequestStartedEventHandler) subscribeStakeRequestStarted(ctx context.Context, sink chan<- *contract.MpcManagerStakeRequestStarted, pubKeys [][]byte) error {
//	if eh.sub != nil {
//		eh.sub.Unsubscribe()
//	}
//
//	err := backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
//		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
//		if err != nil {
//			eh.Logger.Error("Failed to create MpcManagerFilterer", []logger.Field{{"error", err}}...)
//			return errors.WithStack(err)
//		}
//
//		newSub, err := filter.WatchStakeRequestStarted(nil, sink, pubKeys)
//		if err != nil {
//			eh.Logger.Error("Failed to watch StakeRequestStarted event", []logger.Field{{"error", err}}...)
//			return errors.WithStack(err)
//		}
//
//		eh.sub = newSub
//		return nil
//	})
//
//	return err
//}
//
//func (eh *StakeRequestStartedEventHandler) watchStakeRequestStarted(ctx context.Context) {
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-eh.done:
//				return
//			case evt := <-eh.sink:
//				evtObj := dispatcher.NewRootEventObject("StakeRequestStartedEventHandler", evt, ctx)
//				eh.Publisher.Publish(ctx, evtObj)
//			case err := <-eh.sub.Err():
//				eh.Logger.ErrorOnError(err, "Got an error during watching StakeRequestStarted event", []logger.Field{{"error", err}}...)
//			}
//		}
//	}()
//}
