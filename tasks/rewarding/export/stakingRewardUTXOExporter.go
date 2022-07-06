package export

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/signer"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"sync"
	"time"
)

type Cache interface {
	cache.NormalizedParticipantKeysGetter
	cache.GeneratedPubKeyInfoGetter
	cache.IsParticipantChecker
}

// Accept event: *events.StakingTaskDoneEvent
// Accept event: *events.RewardUTXOsFetchedEvent
// Accept event: *events.ExportRewardRequestStartedEvent

// Emit event: *events.RewardExportedEvent

type StakingRewardUTXOExporter struct {
	Logger logger.Logger
	chain.NetworkContext

	MyPubKeyHashHex string

	Publisher dispatcher.Publisher
	SignDoner core.SignDoner

	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.PChainIssuer

	Cache Cache

	once sync.Once
	lock sync.Mutex

	// todo: consider include this field in Cache or dispatcher
	stakingTaskDoneEvtObj            map[string]*dispatcher.EventObject
	utxoFetchedEvtObjMap             map[string]*dispatcher.EventObject
	exportRewardRequestStartedEvtObj map[string]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *StakingRewardUTXOExporter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.stakingTaskDoneEvtObj = make(map[string]*dispatcher.EventObject)
		eh.utxoFetchedEvtObjMap = make(map[string]*dispatcher.EventObject)
		eh.exportRewardRequestStartedEvtObj = make(map[string]*dispatcher.EventObject)
		go func() {
			eh.exportRewardUTXOs(ctx)
		}()
	})

	<-time.After(time.Millisecond * 200) // wait for internal data structures get initialized todo: remove it and use improved strategy.

	switch evt := evtObj.Event.(type) {
	case *events.StakingTaskDoneEvent:
		eh.lock.Lock()
		eh.stakingTaskDoneEvtObj[evt.AddDelegatorTxID.Hex()] = evtObj
		eh.lock.Unlock()
	case *events.RewardUTXOsFetchedEvent:
		eh.lock.Lock()
		eh.utxoFetchedEvtObjMap[evt.AddDelegatorTxID.Hex()] = evtObj
		eh.lock.Unlock()
	case *events.ExportRewardRequestStartedEvent:
		if eh.Cache.IsParticipant(eh.MyPubKeyHashHex, evt.PublicKeyHash.Hex(), evt.ParticipantIndices) {
			eh.lock.Lock()
			eh.exportRewardRequestStartedEvtObj[evt.AddDelegatorTxID.Hex()] = evtObj
			eh.lock.Unlock()
		}
	}
}

func (eh *StakingRewardUTXOExporter) exportRewardUTXOs(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			eh.lock.Lock()
			for txID, evtObj := range eh.exportRewardRequestStartedEvtObj {
				rewardEvt := evtObj.Event.(*events.ExportRewardRequestStartedEvent)
				partiKeys, err := eh.Cache.GetNormalizedParticipantKeys(rewardEvt.PublicKeyHash, rewardEvt.ParticipantIndices)
				if err != nil {
					eh.Logger.Error("StakingRewardUTXOExporter failed to export reward", []logger.Field{
						{"error", err},
						{"exportRewardRequestStartedEvent", rewardEvt},
						{}}...)
					break
				}

				pubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(rewardEvt.PublicKeyHash.Hex())

				stakingTaskDoneEvt := eh.stakingTaskDoneEvtObj[txID].Event.(*events.StakingTaskDoneEvent)
				utxoFetchedEvt := eh.utxoFetchedEvtObjMap[txID].Event.(*events.RewardUTXOsFetchedEvent)

				args := Args{
					NetworkID: eh.NetworkID(),
					//PChainID // todo:
					CChainID:    eh.CChainID(),
					PChainAddr:  stakingTaskDoneEvt.PChainAddress,
					CChainArr:   stakingTaskDoneEvt.CChainAddress,
					RewardUTXOs: utxoFetchedEvt.RewardUTXOs,

					SignDoner: eh.SignDoner,
					SignReqArgs: &signer.SignRequestArgs{
						TaskID:                    txID,
						NormalizedParticipantKeys: partiKeys,
						PubKeyHex:                 pubKeyInfo.GenPubKeyHex,
					},

					CChainIssueClient: eh.CChainIssueClient,
					PChainIssueClient: eh.PChainIssueClient,
				}

				ids, err := exportReward(ctx, &args)
				if err != nil {
					if err != nil {
						eh.Logger.Error("StakingRewardUTXOExporter failed to export reward", []logger.Field{
							{"error", err},
							{"exportRewardRequestStartedEvent", rewardEvt},
							{}}...)
						break
					}
				}

				newEvt := &events.RewardExportedEvent{
					AddDelegatorTxID: rewardEvt.AddDelegatorTxID,
					ExportedTxID:     ids[0],
					ImportedTxID:     ids[1],
				}
				eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "StakingRewardUTXOExporter", newEvt, evtObj.Context))
				delete(eh.exportRewardRequestStartedEvtObj, txID)
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
