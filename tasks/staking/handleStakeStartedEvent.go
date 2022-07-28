package staking

import (
	"context"
	"fmt"
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
	"strings"
	"sync"
	"sync/atomic"
)

const (
	stakeTaskIDPrefix = "STAKE-SIGN-TASK-"
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

	pendingIssueTasksEvtOjbs sync.Map
	pendingIssueTasksCache   sync.Map
	pendingIssueTasks        uint32

	hasIssued   uint32 // 0: no 1: yes
	issuedNonce uint64

	doneStakeTasks uint64
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

		partiKeys, err := eh.getNormalizedPartiKeys(evt.PublicKey, evt.ParticipantIndices)
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to get normalized participant keys")
			return
		}

		taskID := stakeTaskIDPrefix + evt.Raw.TxHash.Hex()

		signRequester := &SignRequester{
			SignDoner: eh.SignDoner,
			SignRequestArgs: SignRequestArgs{
				TaskID:                 taskID,
				CompressedPartiPubKeys: partiKeys,
				CompressedGenPubKeyHex: eh.genPubKeyInfo.CompressedGenPubKeyHex,
			},
		}

		eh.Logger.Debug("Nonce fetched", []logger.Field{
			{"requestID", evt.RequestId.Uint64()},
			{"nonce", nonce},
			{"taskID", evt.Raw.TxHash.Hex()}}...)

		taskCreator := StakeTaskCreator{
			TaskID:                        taskID,
			MpcManagerStakeRequestStarted: evt,
			NetworkContext:                eh.NetworkContext,
			PubKeyHex:                     eh.genPubKeyInfo.CompressedGenPubKeyHex,
			Nonce:                         nonce,
		}
		stakeTask, err := taskCreator.CreateStakeTask()
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to create stake task")
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

		eh.issueStakeTask(ctx, evtObj, stakeTaskWrapper)

	loop:
		nextNonce := atomic.LoadUint64(&eh.issuedNonce) + 1
		stakeTaskWrapperVal, ok := eh.pendingIssueTasksCache.LoadAndDelete(nextNonce)
		if !ok {
			return
		}
		stakeTaskWrapper = stakeTaskWrapperVal.(*StakeTaskWrapper)

		evtObjVal, _ := eh.pendingIssueTasksEvtOjbs.LoadAndDelete(nextNonce)
		evtObj := evtObjVal.(*dispatcher.EventObject)

		eh.issueStakeTask(ctx, evtObj, stakeTaskWrapper)
		goto loop
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
		return errors.WithStack(&ErrTypNonceRegress{ErrMsg: fmt.Sprintf("%v:issuedNonce:%v,givenNonce:%v", ErrMsgNonceRegress, issuedNonce, evmInput.Nonce)})
	}

	addressNonce, err := eh.ChainNoncer.NonceAt(ctx, evmInput.Address, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to request nonce")
	}

	if addressNonce > evmInput.Nonce {
		return errors.WithStack(&ErrTypNonceRegress{ErrMsg: fmt.Sprintf("%v:addressNonce:%v,givenNonce:%v", ErrMsgNonceRegress, addressNonce, evmInput.Nonce)})
	}

	if addressNonce < evmInput.Nonce {
		return errors.WithStack(&ErrTypeNonceJump{ErrMsg: fmt.Sprintf("%v:addressNonce:%v,givenNonce:%v", ErrMsgNonceJump, addressNonce, evmInput.Nonce)})
	}

	return nil
}

func (eh *StakeRequestStartedEventHandler) issueStakeTask(ctx context.Context, evtObj *dispatcher.EventObject, stw *StakeTaskWrapper) {
	stakeTask := stw.StakeTask
	nonce := stakeTask.Nonce
	reqID := stakeTask.RequestID

	if err := eh.checkNonceContinuity(ctx, stakeTask); err != nil {
		switch errors.Cause(err).(type) {
		case *ErrTypNonceRegress:
			eh.Logger.DebugOnError(err, "Stake task CANCELED", []logger.Field{{"stakeTask", stakeTask}}...)
			return
		case *ErrTypeNonceJump:
			eh.pendingIssueTasksCache.Store(nonce, stw)
			eh.pendingIssueTasksEvtOjbs.Store(nonce, evtObj)
			atomic.StoreUint32(&eh.pendingIssueTasks, 1)
			eh.Logger.WarnOnError(err, "Stake task PENDED", []logger.Field{{"pendingIssueStakeTask", atomic.LoadUint32(&eh.pendingIssueTasks)},
				{"stakeTask", stakeTask}}...)
			return
		default:
			eh.Logger.ErrorOnError(err, "Stake task TERMINATED", []logger.Field{{"stakeTask", stakeTask}}...)
			return
		}
	}

	ids, err := stw.IssueTx(ctx)
	signRequester := stw.SignRequester

	if err != nil { // todo: simplify error handling
		switch errors.Cause(err).(type) { // todo: exploring more concrete error types
		case *chain.ErrTypSharedMemoryNotFound:
			eh.Logger.DebugOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
		case *chain.ErrTypInsufficientFunds:
			eh.Logger.DebugOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
		case *chain.ErrTypInvalidNonce:
			eh.Logger.DebugOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
		case *chain.ErrTypConflictAtomicInputs:
			eh.Logger.DebugOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
		case *chain.ErrTypTxHasNoImportedInputs:
			eh.Logger.DebugOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
		case *chain.ErrTypConsumedUTXONotFound:
			eh.Logger.DebugOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
		case *chain.ErrTypNotFound:
			eh.Logger.DebugOnError(err, "Stake task not done", []logger.Field{{"stakeTask", stakeTask}}...)
		default:
			eh.Logger.ErrorOnError(err, "Failed to perform stake task", []logger.Field{{"stakeTask", stakeTask}}...)
		}
		return
	}

	atomic.AddUint64(&eh.doneStakeTasks, 1)

	newEvt := events.StakingTaskDoneEvent{
		TaskID: common.HexToHash(strings.TrimPrefix(stakeTask.TaskID, stakeTaskIDPrefix)),

		ExportTxID:       ids[0],
		ImportTxID:       ids[1],
		AddDelegatorTxID: ids[2],

		RequestID:   stakeTask.RequestID,
		DelegateAmt: stakeTask.DelegateAmt,
		StartTime:   stakeTask.StartTime,
		EndTime:     stakeTask.EndTime,
		NodeID:      stakeTask.NodeID,

		PubKeyHex:     signRequester.CompressedGenPubKeyHex,
		CChainAddress: stakeTask.CChainAddress,
		PChainAddress: stakeTask.PChainAddress,
		Nonce:         stakeTask.Nonce,

		ParticipantPubKeys: signRequester.CompressedPartiPubKeys,
	}
	eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "StakeRequestStartedEventHandler", &newEvt, evtObj.Context))
	eh.Logger.Info("Stake task DONE", []logger.Field{{"doneStakeTasks", atomic.LoadUint64(&eh.doneStakeTasks)}, {"StakingTaskDoneEvent", newEvt}}...)

	atomic.StoreUint64(&eh.issuedNonce, nonce)
	atomic.StoreUint32(&eh.hasIssued, 1)
	if ok := eh.Noncer.ResetBase(reqID, nonce); ok {
		eh.Logger.Info("Noncer base reset", []logger.Field{{"baseReqID", reqID},
			{"baseNonce", nonce}, {"gap", nonce - nonce}}...)
	}
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
