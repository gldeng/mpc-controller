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
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	stakeTaskIDPrefix = "STAKE-"
)

type Cache interface {
	cache.MyIndexGetter
	cache.GeneratedPubKeyInfoGetter
	cache.ParticipantKeysGetter
}

// Accept event: *contract.MpcManagerStakeRequestStarted

// Emit event: *events.StakingTaskDoneEvent

type StakeRequestStartedEventHandler struct {
	Balancer          chain.Balancer
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

	issueTxChan chan *issueTx
	once        sync.Once

	genPubKeyInfo *events.GeneratedPubKeyInfo // todo: value may vary for future key-rotation
	myIndex       *big.Int

	cChainAddress *common.Address // todo: value may vary for future key-rotation
	addrLock      sync.Mutex

	issueTxCache     map[uint64]*issueTx
	issueTxCacheLock sync.RWMutex

	nextNonce uint64

	doneStakeTasks     uint64
	errIssueStakeTasks uint64
}

type issueTx struct {
	*StakeTaskWrapper
	*dispatcher.EventObject
}

func (eh *StakeRequestStartedEventHandler) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.issueTxChan = make(chan *issueTx)
		eh.issueTxCache = make(map[uint64]*issueTx)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case issueTx := <-eh.issueTxChan:
					eh.issueTx(ctx, issueTx) // sequentially issue tx
				}
			}
		}()
	})

	switch evtObj.Event.(type) {
	case *contract.MpcManagerStakeRequestStarted: // todo: params validation
		go eh.signTx(ctx, evtObj) // concurrently sign tx
	}
}

func (eh *StakeRequestStartedEventHandler) signTx(ctx context.Context, evtObj *dispatcher.EventObject) {
	evt := evtObj.Event.(*contract.MpcManagerStakeRequestStarted)

	pubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(evt.PublicKey.Hex())
	if pubKeyInfo == nil {
		eh.Logger.Error("No GeneratedPubKeyInfo found")
		return
	}
	eh.genPubKeyInfo = pubKeyInfo // todo: data race protect

	pubKeyHex := eh.genPubKeyInfo.CompressedGenPubKeyHex
	pubkey, err := crypto.UnmarshalPubKeyHex(pubKeyHex)
	if err != nil {
		eh.Logger.ErrorOnError(err, "Failed to unmarshal public key", []logger.Field{{"publicKey", pubkey}}...)
		return
	}

	eh.addrLock.Lock()
	eh.cChainAddress = addrs.PubkeyToAddresse(pubkey)
	cChainAddr := eh.cChainAddress
	eh.addrLock.Unlock()

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

	// params validation before retrieve nonce
	if err := eh.checkBalance(ctx, *cChainAddr, evt.Amount); err != nil {
		eh.Logger.ErrorOnError(err, "Failed to check balance")
		return
	}
	if err := eh.checkStarTime(evt.StartTime.Int64()); err != nil {
		eh.Logger.ErrorOnError(err, "Failed to check stake start time")
	}

	partiKeys, err := eh.getNormalizedPartiKeys(evt.PublicKey, evt.ParticipantIndices)
	if err != nil {
		eh.Logger.ErrorOnError(err, "Failed to get normalized participant keys")
		return
	}

	nonce := eh.Noncer.GetNonce(evt.RequestId.Uint64()) // todo: how should nonce base adjust in case of validation errors among all participants?
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
		eh.Logger.ErrorOnError(err, "Failed to sign Tx", []logger.Field{{"errSignStakeTask", stakeTask}}...)
		return
	}

	issueTx := &issueTx{stakeTaskWrapper, evtObj}
	select {
	case <-ctx.Done():
	case eh.issueTxChan <- issueTx:
	}
}

func (eh *StakeRequestStartedEventHandler) issueTx(ctx context.Context, issueTx *issueTx) {
	// Cache tx for later issue
	eh.issueTxCacheLock.Lock()
	eh.issueTxCache[issueTx.Nonce] = issueTx
	eh.issueTxCacheLock.Unlock()

	// Sync nonce
	if err := eh.syncNonce(ctx); err != nil {
		eh.Logger.ErrorOnError(err, "Failed to sync nonce")
		return
	}

	// Continuous issue tx
	for i := int(eh.nextNonce); i < int(eh.nextNonce)+len(eh.issueTxCache); i++ {
		eh.issueTxCacheLock.RLock()
		issueTx, ok := eh.issueTxCache[uint64(i)]
		eh.issueTxCacheLock.RUnlock()
		if !ok {
			break
		}
		eh.doIssueTx(ctx, issueTx.EventObject, issueTx.StakeTaskWrapper)
		eh.nextNonce++
	}

	eh.issueTxCacheLock.Lock()
	for nonce, _ := range eh.issueTxCache {
		if nonce < eh.nextNonce {
			delete(eh.issueTxCache, nonce)
		}
	}
	eh.issueTxCacheLock.Unlock()

	var nonces []int
	eh.issueTxCacheLock.RLock()
	for nonce, _ := range eh.issueTxCache {
		nonces = append(nonces, int(nonce))
	}
	eh.issueTxCacheLock.RUnlock()
	sort.Ints(nonces)

	eh.Logger.Debug("Stake tasks stats", []logger.Field{
		{"cachedStakeTasks", len(eh.issueTxCache)},
		{"doneStakeTasks", eh.doneStakeTasks},
		{"errIssueStakeTasks", eh.errIssueStakeTasks},
		{"cachedNonces", nonces}}...)
}

func (eh *StakeRequestStartedEventHandler) doIssueTx(ctx context.Context, evtObj *dispatcher.EventObject, stw *StakeTaskWrapper) error {
	stakeTask := stw.StakeTask
	ids, err := stw.IssueTx(ctx)
	signRequester := stw.SignRequester

	if err != nil { // todo: simplify error handling
		switch errors.Cause(err).(type) { // todo: exploring more concrete error types
		case *chain.ErrTypSharedMemoryNotFound:
			eh.Logger.DebugOnError(err, "Stake task not done for "+chain.ErrMsgSharedMemoryNotFound, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		case *chain.ErrTypInsufficientFunds:
			eh.Logger.WarnOnError(err, "Stake task not done for "+chain.ErrMsgInsufficientFunds, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		case *chain.ErrTypInvalidNonce:
			eh.Logger.DebugOnError(err, "Stake task not done for "+chain.ErrMsgInvalidNonce, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		case *chain.ErrTypConflictAtomicInputs:
			eh.Logger.DebugOnError(err, "Stake task not done for "+chain.ErrMsgConflictAtomicInputs, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		case *chain.ErrTypTxHasNoImportedInputs:
			eh.Logger.DebugOnError(err, "Stake task not done for "+chain.ErrMsgTxHasNoImportedInputs, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		case *chain.ErrTypConsumedUTXONotFound:
			eh.Logger.DebugOnError(err, "Stake task not done for "+chain.ErrMsgConsumedUTXOsNotFound, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		case *chain.ErrTypNotFound:
			eh.Logger.DebugOnError(err, "Stake task not done for "+chain.ErrMsgNotFound, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		case *chain.ErrTypStakeStartTimeExpired: // todo: more measures for this kind of error?
			eh.Logger.ErrorOnError(err, "Failed to stake for "+chain.ErrMsgStakeStartTimeExpired, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		default:
			eh.Logger.ErrorOnError(err, "Failed to stake", []logger.Field{{"errIssueStakeTask", stakeTask}}...)
		}
		atomic.AddUint64(&eh.errIssueStakeTasks, 1)
		return err
	}

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

	eh.doneStakeTasks++

	eh.Logger.Info("Stake task DONE", []logger.Field{{"stakingTaskDoneEvent", newEvt}}...)
	return nil
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

func (eh *StakeRequestStartedEventHandler) syncNonce(ctx context.Context) error {
	addr := eh.cChainAddress
	if addr != nil {
		addressNonce, err := eh.ChainNoncer.NonceAt(ctx, *addr, nil)
		if err != nil {
			return errors.WithStack(err)
		}
		eh.nextNonce = addressNonce
	}
	return nil
}

func (eh *StakeRequestStartedEventHandler) checkNonceContinuity(ctx context.Context, task *StakeTask) error {
	exportTx, err := task.GetSignedExportTx()
	if err != nil {
		return errors.Wrapf(err, "failed to get signed export tx")
	}
	evmInput := exportTx.UnsignedAtomicTx.(*evm.UnsignedExportTx).Ins[0]

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

func (eh *StakeRequestStartedEventHandler) checkBalance(ctx context.Context, addr common.Address, stakeAmt *big.Int) error {
	bl, err := eh.Balancer.BalanceAt(ctx, addr, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	switch bl.Cmp(stakeAmt) {
	case -1:
		fallthrough
	case 0: // take gas fee into account
		return errors.WithStack(&chain.ErrTypInsufficientFunds{
			ErrMsg: fmt.Sprintf("Insufficient funds for stake. address:%v, balance:%v, stakeAmt:%v", addr, bl, stakeAmt)})
	default:
		return nil
	}
}

func (eh *StakeRequestStartedEventHandler) checkStarTime(startTime int64) error {
	startTime_ := time.Unix(startTime, 0)
	now := time.Now()
	switch {
	case startTime_.Before(now):
		return errors.WithStack(&chain.ErrTypStakeStartTimeExpired{})
	case startTime_.Add(time.Second * 5).Before(now):
		eh.Logger.Warn("Stake start time will expired in 5 seconds, continue to issue may fail")
		//return errors.WithStack(&ErrTypStakeStartTimeWillExpireIn5Seconds{})
	case startTime_.Add(time.Second * 10).Before(now):
		eh.Logger.Warn("Stake start time will expired in 10 seconds, continue to issue may fail")
		//return errors.WithStack(&ErrTypStakeStartTimeWillExpireIn10Seconds{})
	case startTime_.Add(time.Second * 20).Before(now):
		eh.Logger.Warn("Stake start time will expired in 20 seconds, continue to issue may fail")
		//return errors.WithStack(&ErrTypStakeStartTimeWillExpireIn20Seconds{})
	case startTime_.Add(time.Second * 40).Before(now):
		eh.Logger.Warn("Stake start time will expired in 40 seconds, continue to issue may fail")
		//return errors.WithStack(&ErrTypStakeStartTimeWillExpireIn40Seconds{})
	case startTime_.Add(time.Second * 60).Before(now):
		eh.Logger.Warn("Stake start time will expired in 60 seconds, continue to issue may fail")
		//return errors.WithStack(&ErrTypStakeStartTimeWillExpireIn60Seconds{})
	}
	return nil
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
