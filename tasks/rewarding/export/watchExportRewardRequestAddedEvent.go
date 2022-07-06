package export

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
// Accept event: *events.GeneratedPubKeyInfoStoredEvent

// Emit event: *contract.ExportRewardRequestAddedEvent

type ExportRewardRequestAddedEventWatcher struct {
	Logger logger.Logger

	ContractAddr common.Address

	Publisher dispatcher.Publisher

	pubKeyBytes [][]byte
	filterer    bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerExportRewardRequestAdded
	done chan struct{}
}

func (eh *ExportRewardRequestAddedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
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
		eh.doWatchExportRewardRequestAdded(evtObj.Context)
	}
}

func (eh *ExportRewardRequestAddedEventWatcher) doWatchExportRewardRequestAdded(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerExportRewardRequestAdded)
	err := eh.subscribeExportRewardRequestAdded(ctx, newSink, eh.pubKeyBytes)
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.watchExportRewardRequestAdded(ctx)
	}
}

func (eh *ExportRewardRequestAddedEventWatcher) subscribeExportRewardRequestAdded(ctx context.Context, sink chan<- *contract.MpcManagerExportRewardRequestAdded, pubKeys [][]byte) error {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			eh.Logger.Error("Failed to create MpcManagerFilterer", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		newSub, err := filter.WatchExportRewardRequestAdded(nil, sink, pubKeys)
		if err != nil {
			eh.Logger.Error("Failed to watch ExportRewardRequestStarted event", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		eh.sub = newSub
		return nil
	})

	return err
}

func (eh *ExportRewardRequestAddedEventWatcher) watchExportRewardRequestAdded(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-eh.done:
				return
			case evt := <-eh.sink:
				transformedEvt := events.ExportRewardRequestAddedEvent{
					AddDelegatorTxID: evt.RewaredStakeTxId,
					PublicKeyHash:    evt.PublicKey,
					TxHash:           evt.Raw.TxHash,
				}
				evtObj := dispatcher.NewRootEventObject("ExportRewardRequestAddedEventWatcher", transformedEvt, ctx)
				eh.Publisher.Publish(ctx, evtObj)
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching ExportRewardRequestAdded event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
