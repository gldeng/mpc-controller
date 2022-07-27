package staking

import (
	"context"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"sync/atomic"
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
	ChainNoncer       chain.Noncer
	Logger            logger.Logger
	MyPubKeyHashHex   string
	Noncer            noncer.Noncer
	PChainIssueClient chain.PChainIssuer
	Publisher         dispatcher.Publisher
	SignDoner         core.SignDoner
	chain.NetworkContext

	genPubKeyInfo *events.GeneratedPubKeyInfo // todo: value may vary for future key-rotation
	myIndex       *big.Int

	hasIssued   uint32 // 0: no 1: yes
	issuedNonce uint64
}

func (eh *StakeRequestStartedEventHandler) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestStarted:
		pubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(evt.PublicKey.Hex())
		if pubKeyInfo == nil {
			eh.Logger.Error("No GeneratedPubKeyInfo found")
			return
		}
		eh.genPubKeyInfo = pubKeyInfo // todo: data race protect
		index := eh.Cache.GetMyIndex(eh.MyPubKeyHashHex, evt.PublicKey.Hex())
		if index == nil {
			eh.Logger.Error("Not found my index.")
			return
		}
		eh.myIndex = index // todo: data race protect

		ok := eh.isParticipant(evt)
		if !ok {
			eh.Logger.Debug("Not participant in *contract.MpcManagerStakeRequestStarted event", []logger.Field{
				{"requestId", evt.RequestId},
				{"TxHash", evt.Raw.TxHash}}...)
			return
		}

		nonce := eh.Noncer.GetNonce(evt.RequestId.Uint64())

		//nonce, err := eh.getNonce(evtObj.Context)
		//if err != err {
		//	eh.Logger.Error("Failed to get nonce", []logger.Field{{"error", err}}...)
		//	return
		//}

		partiKeys, err := eh.getNormalizedPartiKeys(evt.PublicKey, evt.ParticipantIndices)
		if err != nil {
			eh.Logger.Error("Failed to get normalized participant keys", []logger.Field{{"error", err}}...)
			return
		}

		taskID := "STAKE-SIGN-TASK-" + evt.Raw.TxHash.Hex()

		signRequester := &SignRequester{
			SignDoner: eh.SignDoner,
			SignRequestArgs: SignRequestArgs{
				TaskID:                 taskID,
				CompressedPartiPubKeys: partiKeys,
				CompressedGenPubKeyHex: eh.genPubKeyInfo.CompressedGenPubKeyHex,
			},
		}

		eh.Logger.Debug("Got nonce for stake task", []logger.Field{
			{"nonce", nonce},
			{"requestID", evt.RequestId.Uint64()},
			{"taskID", signRequester.SignRequestArgs.TaskID}}...)

		taskCreator := StakeTaskCreator{
			TaskID:                        taskID,
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
			eh.Logger.ErrorOnError(err, "Failed to sign Tx", []logger.Field{{"stakeTask", stakeTask}}...)
			return
		}

		if err := eh.checkNonceContinuity(ctx, stakeTask); err != nil {
			eh.Logger.ErrorOnError(err, "Stake task not allowed to issue", []logger.Field{{"stakeTask", stakeTask}}...)
			return
		}

		ids, err := stakeTaskWrapper.IssueTx(evtObj.Context)
		if err != nil {
			switch errors.Cause(err).(type) { // todo: exploring more concrete error types
			case *chain.ErrTypSharedMemoryNotFound:
				eh.Logger.ErrorOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
			case *chain.ErrTypInsufficientFunds:
				eh.Logger.ErrorOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
			case *chain.ErrTypInvalidNonce:
				eh.Logger.WarnOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
			case *chain.ErrTypConflictAtomicInputs:
				eh.Logger.WarnOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
			case *chain.ErrTypTxHasNoImportedInputs:
				eh.Logger.WarnOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
			case *chain.ErrTypConsumedUTXONotFound:
				eh.Logger.WarnOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
			case *chain.ErrTypNotFound:
				eh.Logger.WarnOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
			default:
				eh.Logger.ErrorOnError(err, "Failed to perform stake task", []logger.Field{{"stakeTask", stakeTask}}...)
			}
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
		eh.Logger.Info("Staking task done", logger.Field{"StakingTaskDoneEvent", newEvt})
		atomic.StoreUint64(&eh.issuedNonce, nonce)
		atomic.StoreUint32(&eh.hasIssued, 1)
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

func (eh *StakeRequestStartedEventHandler) checkNonceContinuity(ctx context.Context, task *StakeTask) error {
	exportTx, err := task.GetSignedExportTx()
	if err != nil {
		return errors.Wrapf(err, "failed to get signed export tx")
	}
	evmInput := exportTx.UnsignedAtomicTx.(*evm.UnsignedExportTx).Ins[0]

	issuedNonce := atomic.LoadUint64(&eh.issuedNonce)
	if atomic.LoadUint32(&eh.hasIssued) == 1 && issuedNonce >= evmInput.Nonce {
		return errors.Errorf("regressed nonce not allowed. issuedNonce: %v, givenNonce: %v", issuedNonce, evmInput.Nonce)
	}

	addressNonce, err := eh.ChainNoncer.NonceAt(ctx, evmInput.Address, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to request nonce")
	}
	if addressNonce != evmInput.Nonce {
		return errors.Errorf("given nonce not equal to address nonce. addressNonce: %v, givenNonce: %v", addressNonce, evmInput.Nonce)
	}
	return nil
}

//func (eh *StakeRequestStartedEventHandler) getNonce(ctx context.Context) (uint64, error) {
//	pubkey, err := crypto.UnmarshalPubKeyHex(eh.genPubKeyInfo.CompressedGenPubKeyHex)
//	if err != nil {
//		return 0, errors.WithStack(err)
//	}
//
//	address := addrs.PubkeyToAddresse(pubkey)
//
//	nonce, err := eh.ChainNoncer.NonceAt(ctx, *address, nil)
//	if err != nil {
//		return 0, errors.WithStack(err)
//	}
//	return nonce, nil
//}

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
