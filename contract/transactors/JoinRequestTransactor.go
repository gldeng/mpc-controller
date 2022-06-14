package transactors

import (
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"time"
)

import (
	"context"
)

// Trigger event: *events.JoinRequestEvent
// Emit event: *events.JoinedRequestEvent

type MpcManagerTransactor struct {
	Logger logger.Logger

	Contract contract.TransactorJoinRequest
	Chain    chain.TransactionReceipt

	Publisher dispatcher.Publisher
}

func (eh *MpcManagerTransactor) Do(evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.JoinRequestEvent:
		eh.joinRequest(evtObj.Context, evt, evtObj)
	}
}

func (eh *MpcManagerTransactor) joinRequest(ctx context.Context, evt *events.JoinRequestEvent, evtObj *dispatcher.EventObject) {
	tx, err := eh.Contract.JoinRequest(ctx, evt.RequestId, evt.Index)
	if err != nil {
		eh.Logger.Error("Failed to join request", []logger.Field{
			{"error", err},
			{"requestId", evt.RequestId},
			{"index", evt.Index}}...)
		return
	}

	rCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err = backoff.RetryFnExponentialForever(eh.Logger, rCtx, func() error {
		rcpt, err := eh.Chain.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return errors.WithStack(err)
		}

		if rcpt.Status != 1 {
			return errors.New("Transaction failed")
		}

		return nil
	})

	eh.Logger.ErrorOnError(err, "Transaction failed", []logger.Field{{"error", err}}...)

	eCtx := context.WithValue(evtObj.Context, "tx", tx.Hash().Hex())

	newEvt := events.JoinedRequestEvent{
		TxHashHex: tx.Hash().Hex(),
		RequestId: evt.RequestId,
		Index:     evt.Index,
	}

	if err == nil {
		eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "contract/transactors/MpcManagerTransactor", newEvt, eCtx))
	}
}
