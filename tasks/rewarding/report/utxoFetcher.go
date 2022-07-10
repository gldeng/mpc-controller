package report

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	// todo: to tune and make it long enough for consequential exporting reward process completed
	// to avoid double check on the same UTXO data set.
	checkUTXOInterval = time.Second * 120
)

// Accept event: *events.ReportedGenPubKeyEvent

// Emit event: *events.UTXOsFetchedEvent

type UTXOFetcher struct {
	Logger       logger.Logger
	PChainClient platformvm.Client
	Publisher    dispatcher.Publisher

	once               sync.Once
	lock               sync.Mutex
	genPubKeyEvtObjMap map[ids.ShortID]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *UTXOFetcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ReportedGenPubKeyEvent:
		eh.once.Do(func() {
			eh.genPubKeyEvtObjMap = make(map[ids.ShortID]*dispatcher.EventObject)
			go func() {
				eh.checkUTXOs(ctx)
			}()
		})

		eh.lock.Lock()
		eh.genPubKeyEvtObjMap[evt.PChainAddress] = evtObj
		eh.lock.Unlock()
	}
}

func (eh *UTXOFetcher) checkUTXOs(ctx context.Context) {
	t := time.NewTicker(checkUTXOInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			eh.lock.Lock()
			for addr, evt := range eh.genPubKeyEvtObjMap {
				utxos, err := eh.getNativeUTXOs(ctx, addr)
				if err != nil {
					eh.Logger.Error("Failed to get native UTXOss", []logger.Field{{"error", err}}...)
					continue
				}

				//var rewardedTxIDs = make(map[ids.ID]struct{})
				//for _, utxo := range utxos {
				//	rewardedTxIDs[utxo.TxID] = struct{}{}
				//}
				//for txID, _ := range rewardedTxIDs {
				//	rewardUTXO, err := eh.getRewardUTXOs(ctx, txID)
				//	if err != nil {
				//		eh.Logger.Error("Failed to get reward UTXOs", []logger.Field{{"error", err}}...)
				//		delete(rewardedTxIDs, txID)
				//		continue
				//	}
				//	if rewardUTXO == nil {
				//		delete(rewardedTxIDs, txID)
				//		continue
				//	}
				//}
				//
				//if len(rewardedTxIDs) == 0 {
				//	eh.Logger.Debug("Found no reward UTXOs", []logger.Field{{"pChainAddress", addr}}...)
				//	continue
				//}
				//
				//var principalUTXOs []*avax.UTXO
				//var rewardUTXOs []*avax.UTXO
				//for _, utxo := range utxos {
				//	_, ok := rewardedTxIDs[utxo.TxID]
				//	if ok {
				//		switch utxo.OutputIndex {
				//		case 0:
				//			principalUTXOs = append(principalUTXOs, utxo)
				//		case 1:
				//			rewardUTXOs = append(rewardUTXOs, utxo)
				//		}
				//	}
				//}

				if len(utxos) == 0 {
					eh.Logger.Debug("Found no native UTXOs", []logger.Field{{"pChainAddress", addr}}...)
					continue
				}

				mpcUTXOs := myAvax.MpcUTXOsFromUTXOs(utxos)

				newEvt := &events.UTXOsFetchedEvent{
					NativeUTXOs: utxos,
					MpcUTXOs:    mpcUTXOs,
				}

				copier.Copy(&newEvt, evt.Event.(*events.ReportedGenPubKeyEvent))

				eh.Publisher.Publish(ctx, dispatcher.NewRootEventObject("UTXOFetcher", newEvt, ctx))
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (eh *UTXOFetcher) getNativeUTXOs(ctx context.Context, addr ids.ShortID) ([]*avax.UTXO, error) {
	var utxoBytesArr [][]byte
	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		var err error
		utxoBytesArr, _, _, err = eh.PChainClient.GetUTXOs(ctx, []ids.ShortID{addr}, 0, addr, ids.ID{})
		if err != nil {
			return errors.Wrap(err, "failed to request native UTXOs")
		}
		return nil
	})

	utxos, err := myAvax.ParseUTXOs(utxoBytesArr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse native UTXO bytes")
	}

	var utxosFiltered []*avax.UTXO
	for _, utxo := range utxos {
		if utxo != nil {
			utxosFiltered = append(utxosFiltered, utxo)
		}
	}
	return utxosFiltered, nil
}

//func (eh *UTXOFetcher) getRewardUTXOs(ctx context.Context, txID ids.ID) ([]*avax.UTXO, error) {
//	var utxoBytesArr [][]byte
//	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
//		var err error
//		utxoBytesArr, err = eh.PChainClient.GetRewardUTXOs(ctx, &api.GetTxArgs{TxID: txID})
//		if err != nil {
//			return errors.Wrap(err, "failed to request reward UTXOs")
//		}
//		return nil
//	})
//
//	utxos, err := myAvax.ParseUTXOs(utxoBytesArr)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to parse UTXO bytes")
//	}
//
//	var utxosFiltered []*avax.UTXO
//	for _, utxo := range utxos {
//		if utxo != nil && utxo.OutputIndex == 1 { // output index: 0-for principal, 1-for delegator, 2-for-validator
//			utxosFiltered = append(utxosFiltered, utxo)
//		}
//	}
//	return utxosFiltered, nil
//}
