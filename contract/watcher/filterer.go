package watcher

import (
	"github.com/avalido/mpc-controller/contract"
)

type Filterer struct {
	contract.MpcManagerFilterer
}

// Subscribe event: *events.ContractFiltererCreated
// Subscribe event:  *events.GeneratedPubKeyInfoStored

// Process event:  *contract.ParticipantAdded
// Process event:  *contract.KeygenRequestAdded
// Process event:  *contract.KeyGenerated
// Process event:  *contract.StakeRequestAdded
// Process event:  *contract.RequestStarted
//
//type StakeRequestAddedEventWatcher struct {
//	L       logger.L
//	ContractAddr common.Address
//	P    dispatcher.P
//	pubKeyBytes  [][]byte
//	filterer     bind.ContractFilterer
//
//	sub  event.Subscription
//	sink chan *contract.MpcManagerStakeRequestAdded
//	done chan struct{}
//
//	hasPublishedReq bool
//	lastReqID       uint64
//
//	participantAddedSink chan *contract.MpcManagerParticipantAdded
//	keygenRequestAddedSink chan *contract.MpcManagerKeygenRequestAdded
//	keyGeneratedEventSink chan *contract.MpcManagerKeyGenerated
//	stakeRequestAddedSink chan *contract.MpcManagerStakeRequestAdded
//	requestStartedSink chan *contract.MpcManagerRequestStarted
//}
//
//func (eh *StakeRequestAddedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
//	switch evt := evtObj.Event.(type) {
//	case *events.ContractFiltererCreated:
//		eh.filterer = evt.Filterer
//	case *events.GeneratedPubKeyInfoStored:
//		dnmPubKeyBtes, err := crypto.DenormalizePubKeyFromHex(evt.Val.CompressedGenPubKeyHex)
//		if err != nil {
//			eh.L.Error("Failed to denormalized generated public key", []logger.Field{{"error", err}}...)
//			break
//		}
//
//		eh.pubKeyBytes = append(eh.pubKeyBytes, dnmPubKeyBtes)
//	}
//	if len(eh.pubKeyBytes) > 0 {
//		eh.doWatchStakeRequestAdded(ctx)
//	}
//}
//
//func (eh *StakeRequestAddedEventWatcher) doWatchStakeRequestAdded(ctx context.Context) {
//	newSink := make(chan *contract.MpcManagerStakeRequestAdded)
//	err := eh.subscribeStakeRequestAdded(ctx, newSink, eh.pubKeyBytes)
//	if err == nil {
//		eh.sink = newSink
//		if eh.done != nil {
//			close(eh.done)
//		}
//		eh.done = make(chan struct{})
//		eh.watchStakeRequestAdded(ctx)
//	}
//}
//
//func (eh *StakeRequestAddedEventWatcher) subscribeStakeRequestAdded(ctx context.Context, sink chan<- *contract.MpcManagerStakeRequestAdded, pubKeys [][]byte) error {
//	if eh.sub != nil {
//		eh.sub.Unsubscribe()
//	}
//
//	err := backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
//		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//
//		newSub, err := filter.WatchStakeRequestAdded(nil, sink, pubKeys)
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//		eh.sub = newSub
//		return false, nil
//	})
//	err = errors.Wrapf(err, "failed to subscribe StakeRequestAdded event")
//	return err
//}
//
//func (eh *StakeRequestAddedEventWatcher) watchStakeRequestAdded(ctx context.Context) {
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-eh.done:
//			return
//		case <-eh.participantAddedSink:
//		case <-eh.keygenRequestAddedSink:
//		case <-eh.keyGeneratedEventSink:
//		case <-eh.stakeRequestAddedSink:
//		case <-eh.requestStartedSink:
//
//		case evt := <-eh.sink:
//			eh.L.DebugOnTrue(eh.hasPublishedReq && evt.RequestId.Uint64() != eh.lastReqID+1,
//				"Received un-continuous emitted stake request",
//				[]logger.Field{{"expectedReqID", eh.lastReqID + 1},
//					{"receivedReqID", evt.RequestId.Uint64()}}...)
//
//			evtObj := dispatcher.NewEvtObj(evt, nil)
//			eh.P.Publish(ctx, evtObj)
//			eh.L.Debug("StakeRequestAdded emitted", []logger.Field{{"reqID", evt.RequestId}, {"txHash", evt.Raw.TxHash}}...)
//			eh.hasPublishedReq = true
//			eh.lastReqID = evt.RequestId.Uint64()
//			prom.StakeRequestAdded.Inc()
//		case err := <-eh.participantAddedSub.Err():
//		case err := <-eh.keygenRequestAddedSub.Err():
//		case err := <-eh.keyGeneratedEventSub.Err():
//		case err := <-eh.stakeRequestAddedEventSub.Err():
//		case err := <-eh.requestStartedEventSub.Err():
//
//		case err := <-eh.sub.Err():
//			eh.L.ErrorOnError(err, "Got an error during watching StakeRequestAdded event")
//		}
//	}
//}
