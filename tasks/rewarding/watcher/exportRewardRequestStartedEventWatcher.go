package watcher

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

// Emit event: *contract.ExportRewardRequestStartedEvent

type ExportRewardRequestStartedEventWatcher struct {
	Logger logger.Logger

	ContractAddr common.Address

	Publisher dispatcher.Publisher

	pubKeyBytes [][]byte
	filterer    bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerExportRewardRequestStarted
	done chan struct{}
}

func (eh *ExportRewardRequestStartedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
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
		eh.doWatchExportRewardRequestStarted(evtObj.Context)
	}
}

func (eh *ExportRewardRequestStartedEventWatcher) doWatchExportRewardRequestStarted(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerExportRewardRequestStarted)
	err := eh.subscribeExportRewardRequestStarted(ctx, newSink, eh.pubKeyBytes)
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.watchExportRewardRequestStarted(ctx)
	}
}

func (eh *ExportRewardRequestStartedEventWatcher) subscribeExportRewardRequestStarted(ctx context.Context, sink chan<- *contract.MpcManagerExportRewardRequestStarted, pubKeys [][]byte) error {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			eh.Logger.Error("Failed to create MpcManagerFilterer", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		newSub, err := filter.WatchExportRewardRequestStarted(nil, sink, pubKeys)
		if err != nil {
			eh.Logger.Error("Failed to watch ExportRewardRequestStarted event", []logger.Field{{"error", err}}...)
			return errors.WithStack(err)
		}

		eh.sub = newSub
		return nil
	})

	return err
}

func (eh *ExportRewardRequestStartedEventWatcher) watchExportRewardRequestStarted(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-eh.done:
				return
			case evt := <-eh.sink:
				transformedEvt := events.ExportRewardRequestStartedEvent{
					AddDelegatorTxID:   evt.RewaredStakeTxId,
					PublicKeyHash:      evt.PublicKey,
					ParticipantIndices: evt.ParticipantIndices,
					TxHash:             evt.Raw.TxHash,
				}
				evtObj := dispatcher.NewRootEventObject("ExportRewardRequestStartedEventWatcher", transformedEvt, ctx)
				eh.Publisher.Publish(ctx, evtObj)
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching ExportRewardRequestStarted event", []logger.Field{{"error", err}}...)
			}
		}
	}()
}
