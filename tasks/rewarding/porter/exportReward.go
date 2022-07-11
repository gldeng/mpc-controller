package porter

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/signer"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto/secp256k1r"
	portIssuer "github.com/avalido/mpc-controller/utils/port/issuer"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// Accept event: *events.StakingTaskDoneEvent
// Accept event: *events.UTXOsFetchedEvent
// Accept event: *events.ExportRewardRequestStartedEvent

// Emit event: *events.RewardExportedEvent

type StakingRewardExporter struct {
	CChainIssueClient chain.CChainIssuer
	Cache             cache.ICache
	Logger            logger.Logger
	MyPubKeyHashHex   string
	PChainIssueClient chain.PChainIssuer
	Publisher         dispatcher.Publisher
	SignDoner         core.SignDoner
	chain.NetworkContext

	once sync.Once
	lock sync.Mutex

	// todo: consider include this field in Cache or dispatcher
	stakingTaskDoneEvtObj            map[string]*dispatcher.EventObject
	utxoFetchedEvtObjMap             map[string]*dispatcher.EventObject
	exportRewardRequestStartedEvtObj map[string]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *StakingRewardExporter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.stakingTaskDoneEvtObj = make(map[string]*dispatcher.EventObject)
		eh.utxoFetchedEvtObjMap = make(map[string]*dispatcher.EventObject)
		eh.exportRewardRequestStartedEvtObj = make(map[string]*dispatcher.EventObject)
		go func() {
			eh.exportRewardUTXOs(ctx)
		}()
	})

	switch evt := evtObj.Event.(type) {
	case *events.StakingTaskDoneEvent:
		eh.lock.Lock()
		eh.stakingTaskDoneEvtObj[evt.AddDelegatorTxID.Hex()] = evtObj
		eh.lock.Unlock()
	case *events.UTXOsFetchedEvent:
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

func (eh *StakingRewardExporter) exportRewardUTXOs(ctx context.Context) {
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
					eh.Logger.Error("StakingRewardExporter failed to export reward", []logger.Field{
						{"error", err},
						{"exportRewardRequestStartedEvent", rewardEvt},
						{}}...)
					break
				}

				pubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(rewardEvt.PublicKeyHash.Hex())

				stakingTaskDoneEvt := eh.stakingTaskDoneEvtObj[txID].Event.(*events.StakingTaskDoneEvent)
				utxoFetchedEvt := eh.utxoFetchedEvtObjMap[txID].Event.(*events.UTXOsFetchedEvent)

				args := Args{
					NetworkID: eh.NetworkID(),
					//PChainID // todo:
					CChainID:    eh.CChainID(),
					PChainAddr:  stakingTaskDoneEvt.PChainAddress,
					CChainArr:   stakingTaskDoneEvt.CChainAddress,
					RewardUTXOs: utxoFetchedEvt.NativeUTXOs,

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
						eh.Logger.Error("StakingRewardExporter failed to export reward", []logger.Field{
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
				eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "StakingRewardExporter", newEvt, evtObj.Context))
				delete(eh.exportRewardRequestStartedEvtObj, txID)
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

type Args struct {
	NetworkID   uint32
	PChainID    ids.ID
	CChainID    ids.ID
	PChainAddr  ids.ShortID
	CChainArr   common.Address
	RewardUTXOs []*avax.UTXO

	SignDoner   core.SignDoner
	SignReqArgs *signer.SignRequestArgs

	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.PChainIssuer
}

func exportReward(ctx context.Context, args *Args) ([2]ids.ID, error) {
	amountToExport := args.RewardUTXOs[0].Out.(*secp256k1fx.TransferOutput).Amount()
	myExportTxArgs := &pchain.Args{
		NetworkID:          args.NetworkID,
		BlockchainID:       args.PChainID,
		DestinationChainID: args.CChainID,
		Amount:             amountToExport, // todo: to be tuned
		To:                 args.PChainAddr,
		UTXOs:              args.RewardUTXOs,
	}

	myImportTxArgs := &cchain.Args{
		NetworkID:     args.NetworkID,
		BlockchainID:  args.CChainID,
		SourceChainID: args.PChainID,
		To:            args.CChainArr,
		//BaseFee:      *big.Int todo: tobe consider
	}

	myTxs := &Txs{
		UnsignedExportTxArgs: myExportTxArgs,
		UnsignedImportTx:     myImportTxArgs,
	}

	mySigner := &signer.Signer{args.SignDoner, *args.SignReqArgs}
	myVerifier := &secp256k1r.Verifier{PChainAddress: args.PChainAddr}
	myIssuer := &portIssuer.Issuer{args.CChainIssueClient, args.PChainIssueClient, portIssuer.P2C}
	myPorter := porter.Porter{myTxs, mySigner, myIssuer, myVerifier}

	txIds, err := myPorter.SignAndIssueTxs(ctx)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	return txIds, nil
}
