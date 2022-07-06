package report

import (
	"context"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// Accept event: *events.StakingPeriodEndedEvent

// Emit event: *events.RewardUTXOsFetchedEvent

type StakingRewardUTXOFetcher struct {
	Logger logger.Logger

	Publisher dispatcher.Publisher

	RewardUTXOGetter chain.RewardUTXOGetter

	once           sync.Once
	lock           sync.Mutex
	endedEvtObjMap map[string]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *StakingRewardUTXOFetcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.StakingPeriodEndedEvent:
		eh.once.Do(func() {
			eh.endedEvtObjMap = make(map[string]*dispatcher.EventObject)
			go func() {
				eh.fetchRewardUTXOs(ctx)
			}()
		})

		eh.lock.Lock()
		eh.endedEvtObjMap[evt.AddDelegatorTxID.String()] = evtObj
		eh.lock.Unlock()
	}
}

func (eh *StakingRewardUTXOFetcher) fetchRewardUTXOs(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			eh.lock.Lock()
			for txID, evtObj := range eh.endedEvtObjMap {
				evt := evtObj.Event.(*events.StakingPeriodEndedEvent)
				utxos := eh.retryRequestRewardUTXOs(ctx, evt.AddDelegatorTxID)
				newEvt := &events.RewardUTXOsFetchedEvent{
					AddDelegatorTxID: evt.AddDelegatorTxID,
					RewardUTXOs:      utxos,
					PubKeyHex:        evt.PubKeyHex,
				}
				eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "StakingRewardUTXOFetcher", newEvt, evtObj.Context))
				delete(eh.endedEvtObjMap, txID)
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (eh *StakingRewardUTXOFetcher) retryRequestRewardUTXOs(ctx context.Context, txID ids.ID) []*avax.UTXO {
	var results = make([]*avax.UTXO, 0)

	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		utxos, err := eh.requestRewardUTXOs(ctx, txID)
		if err != nil {
			return errors.Wrapf(err, "failed to request reward UTXOs for txID:%v", txID)
		}
		if len(utxos) == 0 {
			return errors.Errorf("no reward UTXO found for txID:%v", txID)
		}
		results = utxos
		return nil
	})

	return results
}

func (eh *StakingRewardUTXOFetcher) requestRewardUTXOs(ctx context.Context, txID ids.ID) ([]*avax.UTXO, error) {
	utxosBytesArr, err := eh.RewardUTXOGetter.GetRewardUTXOs(ctx, &api.GetTxArgs{TxID: txID})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var utxos = make([]*avax.UTXO, len(utxosBytesArr))

	for _, utxoBytes := range utxosBytesArr {
		var utxo avax.UTXO
		_, err := platformvm.Codec.Unmarshal(utxoBytes, &utxo)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		utxos = append(utxos, &utxo)
	}

	return utxos, nil
}
