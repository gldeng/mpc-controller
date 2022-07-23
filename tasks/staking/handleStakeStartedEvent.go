package staking

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
	"time"
)

type Cache interface {
	cache.MyIndexGetter
	cache.GeneratedPubKeyInfoGetter
	cache.ParticipantKeysGetter
}

// Accept event: *contract.MpcManagerStakeRequestStarted

// Emit event: *events.StakingTaskDoneEvent

type StakeRequestStartedEventHandler struct {
	CChainIssueClient chain.CChainIssuer
	Cache             Cache // todo: to use cache.ICache instead
	Logger            logger.Logger
	MyPubKeyHashHex   string
	Noncer            chain.Noncer
	PChainIssueClient chain.PChainIssuer
	Publisher         dispatcher.Publisher
	SignDoner         core.SignDoner
	chain.NetworkContext

	genPubKeyInfo *events.GeneratedPubKeyInfo
	myIndex       *big.Int
}

func (eh *StakeRequestStartedEventHandler) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestStarted:
		pubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(evt.PublicKey.Hex())
		if pubKeyInfo == nil {
			eh.Logger.Error("No GeneratedPubKeyInfo found")
			return
		}
		eh.genPubKeyInfo = pubKeyInfo
		index := eh.Cache.GetMyIndex(eh.MyPubKeyHashHex, evt.PublicKey.Hex())
		if index == nil {
			eh.Logger.Error("Not found my index.")
			return
		}
		eh.myIndex = index

		ok := eh.isParticipant(evt)
		if !ok {
			eh.Logger.Debug("Not participant in *contract.MpcManagerStakeRequestStarted event", []logger.Field{
				{"requestId", evt.RequestId},
				{"TxHash", evt.Raw.TxHash}}...)
			return
		}

		nonce, err := eh.getNonce(evtObj.Context)
		if err != err {
			eh.Logger.Error("Failed to get nonce", []logger.Field{{"error", err}}...)
			return
		}

		partiKeys, err := eh.getNormalizedPartiKeys(evt.PublicKey, evt.ParticipantIndices)
		if err != nil {
			eh.Logger.Error("Failed to get normalized participant keys", []logger.Field{{"error", err}}...)
			return
		}

		signRequester := &SignRequester{
			SignDoner: eh.SignDoner,
			SignRequestArgs: SignRequestArgs{
				TaskID:                 evt.Raw.TxHash.Hex(),
				CompressedPartiPubKeys: partiKeys,
				CompressedGenPubKeyHex: eh.genPubKeyInfo.CompressedGenPubKeyHex,
			},
		}

		taskCreator := StakeTaskCreator{
			MpcManagerStakeRequestStarted: evt,
			NetworkContext:                eh.NetworkContext,
			PubKeyHex:                     eh.genPubKeyInfo.CompressedGenPubKeyHex,
			Nonce:                         nonce,
		}
		stakeTask, err := taskCreator.CreateStakeTask()
		if err != nil {
			eh.Logger.Error("Failed to create stake task", []logger.Field{{"error", err}}...)
			return
		}

		stakeTaskWrapper := &StakeTaskWrapper{
			CChainIssueClient: eh.CChainIssueClient,
			Logger:            eh.Logger,
			PChainIssueClient: eh.PChainIssueClient,
			SignRequester:     signRequester,
			StakeTask:         stakeTask,
		}

		err = stakeTaskWrapper.SignTx(evtObj.Context)
		if err != nil {
			eh.Logger.Error("Failed to sign Tx", []logger.Field{{"error", err}}...)
			return
		}

		n := rand.Intn(10000)
		time.Sleep(time.Millisecond * time.Duration(n)) // reduce concurrent conflict

		ids, err := stakeTaskWrapper.IssueTx(evtObj.Context)
		if err != nil {
			eh.Logger.Error("Failed to perform stake task", []logger.Field{
				{"stakeTask", stakeTask},
				{"error", err}}...)
			return
		}

		newEvt := events.StakingTaskDoneEvent{
			TaskID: evt.Raw.TxHash,

			ExportTxID:       ids[0],
			ImportTxID:       ids[1],
			AddDelegatorTxID: ids[2],

			RequestID:   stakeTask.RequestID,
			DelegateAmt: stakeTask.DelegateAmt,
			StartTime:   stakeTask.StartTime,
			EndTime:     stakeTask.EndTime,
			NodeID:      stakeTask.NodeID,

			PubKeyHex:     eh.genPubKeyInfo.CompressedGenPubKeyHex,
			CChainAddress: stakeTask.CChainAddress,
			PChainAddress: stakeTask.PChainAddress,
			Nonce:         stakeTask.Nonce,

			ParticipantPubKeys: partiKeys,
		}
		eh.Publisher.Publish(evtObj.Context, dispatcher.NewEventObjectFromParent(evtObj, "StakeRequestStartedEventHandler", &newEvt, evtObj.Context))
	}
}

// todo: use cache.IsParticipantChecker
func (eh *StakeRequestStartedEventHandler) isParticipant(req *contract.MpcManagerStakeRequestStarted) bool {
	var participating bool
	for _, index := range req.ParticipantIndices {
		if index.Cmp(eh.myIndex) == 0 {
			participating = true
			break
		}
	}

	if !participating {
		return false
	}

	return true
}

func (eh *StakeRequestStartedEventHandler) getNonce(ctx context.Context) (uint64, error) {
	pubkey, err := crypto.UnmarshalPubKeyHex(eh.genPubKeyInfo.CompressedGenPubKeyHex)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	address := addrs.PubkeyToAddresse(pubkey)

	nonce, err := eh.Noncer.NonceAt(ctx, *address, nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return nonce, nil
}

// todo: use cache.NormalizedParticipantKeysGetter instead
func (eh *StakeRequestStartedEventHandler) getNormalizedPartiKeys(genPubKeyHash common.Hash, partiIndices []*big.Int) ([]string, error) {
	partiKeys := eh.Cache.GetParticipantKeys(genPubKeyHash, partiIndices)
	if partiKeys == nil {
		return nil, errors.New("Found no participant keys.")
	}

	normalized, err := crypto.NormalizePubKeys(partiKeys)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to normalized participant public keys: %v", partiKeys)
	}

	return normalized, nil
}
