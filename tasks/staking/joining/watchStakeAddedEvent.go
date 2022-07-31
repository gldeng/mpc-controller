package joining

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"time"
)

// Subscribe event: *events.ContractFiltererCreatedEvent
// Subscribe event:  *events.GeneratedPubKeyInfoStoredEvent

// Publish event:  *contract.MpcManagerStakeRequestAdded

type StakeRequestAddedEventWatcher struct {
	Logger       logger.Logger
	ContractAddr common.Address
	Publisher    dispatcher.Publisher
	pubKeyBytes  [][]byte
	filterer     bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerStakeRequestAdded
	done chan struct{}

	hasPublishedReq bool
	lastReqID       uint64
}

func (eh *StakeRequestAddedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
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
		eh.doWatchStakeRequestAdded(ctx)
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

	err := backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			return true, errors.WithStack(err)
		}

		newSub, err := filter.WatchStakeRequestAdded(nil, sink, pubKeys)
		if err != nil {
			return true, errors.WithStack(err)
		}
		eh.sub = newSub
		return false, nil
	})
	err = errors.Wrapf(err, "failed to subscribe StakeRequestAdded event")
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
				eh.Logger.WarnOnTrue(eh.hasPublishedReq && evt.RequestId.Uint64() != eh.lastReqID+1,
					"Received un-continuous emitted stake request",
					[]logger.Field{{"expectedReqID", eh.lastReqID + 1},
						{"receivedReqID", evt.RequestId.Uint64()}}...)

				evtObj := dispatcher.NewEvtObj(evt, nil)
				eh.Publisher.Publish(ctx, evtObj)
				eh.Logger.Debug("StakeRequestAddedEvent emitted", []logger.Field{{"reqID", evt.RequestId}, {"txHash", evt.Raw.TxHash}}...)
				eh.hasPublishedReq = true
				eh.lastReqID = evt.RequestId.Uint64()
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching StakeRequestAdded event")
			}
		}
	}()
}

type stakeReqsAddedArr []*contract.MpcManagerStakeRequestAdded

func (s stakeReqsAddedArr) Len() int {
	return len(s)
}

func (s stakeReqsAddedArr) Less(i, j int) bool {
	return s[i].RequestId.Uint64() < s[j].RequestId.Uint64()
}

func (s stakeReqsAddedArr) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
