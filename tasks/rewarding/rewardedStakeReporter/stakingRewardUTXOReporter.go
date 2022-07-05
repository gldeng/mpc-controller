package rewardedStakeReporter

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// Accept event: *events.RewardUTXOsFetchedEvent

// Emit event: *events.RewardUTXOsReportedEvent

type RewardedStakeReporter struct {
	Logger logger.Logger

	Publisher dispatcher.Publisher

	Transactor contract.TransactorReportRewardUTXOs
	Receipter  chain.Receipter

	once                 sync.Once
	lock                 sync.Mutex
	utxoFetchedEvtObjMap map[string]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *RewardedStakeReporter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.RewardUTXOsFetchedEvent:
		eh.once.Do(func() {
			eh.utxoFetchedEvtObjMap = make(map[string]*dispatcher.EventObject)
			go func() {
				eh.reportRewardUTXOs(ctx)
			}()
		})

		eh.lock.Lock()
		eh.utxoFetchedEvtObjMap[evt.AddDelegatorTxID.String()] = evtObj
		eh.lock.Unlock()
	}
}

func (eh *RewardedStakeReporter) reportRewardUTXOs(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			eh.lock.Lock()
			for txID, evtObj := range eh.utxoFetchedEvtObjMap {
				evt := evtObj.Event.(*events.RewardUTXOsFetchedEvent)
				var utxoIDArr []string
				for _, utxo := range evt.RewardUTXOs {
					utxoIDArr = append(utxoIDArr, utxo.String())
				}

				txHash := eh.retryReportRewardUTXOs(ctx, evt.AddDelegatorTxID, utxoIDArr)
				newEvt := &events.RewardUTXOsReportedEvent{
					AddDelegatorTxID: evt.AddDelegatorTxID,
					RewardUTXOIDs:    utxoIDArr,
					TxHash:           txHash,
				}
				eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "RewardedStakeReporter", newEvt, evtObj.Context))
				delete(eh.utxoFetchedEvtObjMap, txID)
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (eh *RewardedStakeReporter) retryReportRewardUTXOs(ctx context.Context, addDelegatorTxID [32]byte, rewardUTXOIDs []string) (txHash *common.Hash) {
	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		tx, err := eh.Transactor.ReportRewardUTXOs(ctx, addDelegatorTxID, rewardUTXOIDs)
		if err != nil {
			return errors.Wrapf(err, "failed to report reward UTXOs for addDelegatorTxID:%v", addDelegatorTxID)
		}

		time.Sleep(time.Second * 3)

		rcpt, err := eh.Receipter.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return errors.WithStack(err)
		}

		if rcpt.Status != 1 {
			return errors.Errorf("reporting reward UTXOs' transaction for addDelegatorTxID %q failed", addDelegatorTxID)
		}

		newTxHash := tx.Hash()
		txHash = &newTxHash
		return nil
	})
	return
}
