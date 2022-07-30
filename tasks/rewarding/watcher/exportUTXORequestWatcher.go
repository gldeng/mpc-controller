package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"time"
)

// Accept event: *events.ContractFiltererCreatedEvent
// Accept event: *events.ReportedGenPubKeyEvent

// Emit event: *contract.ExportUTXORequestEvent

type ExportUTXORequestWatcher struct {
	ContractAddr common.Address
	Logger       logger.Logger
	Publisher    dispatcher.Publisher

	pubKeyBytes [][]byte
	filterer    bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerExportUTXORequest
	done chan struct{}
}

func (eh *ExportUTXORequestWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererCreatedEvent:
		eh.filterer = evt.Filterer
	case *events.ReportedGenPubKeyEvent:
		eh.pubKeyBytes = append(eh.pubKeyBytes, bytes.HexToBytes(evt.GenPubKeyHex))
	}
	if len(eh.pubKeyBytes) > 0 {
		eh.watchExportUTXORequestEvent(ctx)
	}
}

func (eh *ExportUTXORequestWatcher) watchExportUTXORequestEvent(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerExportUTXORequest)
	err := eh.subscribeExportUTXORequestEvent(ctx, newSink, eh.pubKeyBytes)
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.receiveExportUTXORequestEvent(ctx)
	}
}

func (eh *ExportUTXORequestWatcher) subscribeExportUTXORequestEvent(ctx context.Context, sink chan<- *contract.MpcManagerExportUTXORequest, pubKeys [][]byte) error {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}
	err := backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			return true, errors.WithStack(err)
		}

		newSub, err := filter.WatchExportUTXORequest(nil, sink, pubKeys)
		if err != nil {
			return true, errors.WithStack(err)
		}
		eh.sub = newSub
		return false, nil
	})
	err = errors.Wrapf(err, "failed to subscribe ExportUTXORequest event")
	return err
}

func (eh *ExportUTXORequestWatcher) receiveExportUTXORequestEvent(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-eh.done:
				return
			case evt := <-eh.sink:
				transformedEvt := events.ExportUTXORequestEvent{
					TxID:          evt.TxId,
					GenPubKeyHash: evt.GenPubKey,
					TxHash:        evt.Raw.TxHash,
				}
				copier.Copy(&transformedEvt, evt)
				evtObj := dispatcher.NewEvtObj(&transformedEvt, nil)
				eh.Publisher.Publish(ctx, evtObj)
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching ExportRewardRequest event for ExportUTXORequestWatcher")
			}
		}
	}()
}
