package staking

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/avalido/mpc-controller/utils/work"
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

// Subscribe event: *events.RequestStarted

// Publish event:

type StakeRequestStarted struct {
	PubKey []byte

	Balancer          chain.Balancer
	CChainIssueClient chain.CChainIssuer
	Cache             Cache // todo: to use cache.ICache instead
	ChainNoncer       chain.Noncer
	Logger            logger.Logger
	Noncer            noncer.Noncer
	PChainIssueClient chain.PChainIssuer
	Publisher         dispatcher.Publisher
	SignDoner         core.SignDoner
	chain.NetworkContext

	DB         storage.DB
	Transactor transactor.Transactor

	requestStartedChan chan *events.RequestStarted
	signStakeTxWs      *work.Workshop

	issueTxChan chan *issueTx
	once        sync.Once

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
	*storage.StakeRequest
}

func (eh *StakeRequestStarted) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.signStakeTxWs = work.NewWorkshop(eh.Logger, "signStakeTx", time.Minute*10, 10)

		eh.requestStartedChan = make(chan *events.RequestStarted, 1024)

		eh.issueTxChan = make(chan *issueTx)
		eh.issueTxCache = make(map[uint64]*issueTx)

		go eh.signTx(ctx)
		go eh.issueTx(ctx)
		go eh.checkIssueTxCache(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *events.RequestStarted:
		select {
		case <-ctx.Done():
			return
		case eh.requestStartedChan <- evt:
		}
	}
}

func (eh *StakeRequestStarted) signTx(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-eh.requestStartedChan:
			stakeReq := storage.StakeRequest{}
			joinReq := storage.JoinRequest{
				ReqHash: evt.RequestHash,
				Args:    &stakeReq,
			}
			if err := eh.DB.LoadModel(ctx, &joinReq); err != nil {
				eh.Logger.Debug("No JoinRequest load for stake", []logger.Field{{"key", evt.RequestHash}}...)
				break
			}

			if !joinReq.PartiId.Joined(evt.ParticipantIndices) {
				eh.Logger.Debug("Not joined stake request", []logger.Field{{"reqNo", stakeReq.ReqNo}, {"reqHash", evt.RequestHash}}...)
				break
			}

			genPubKey := storage.GeneratedPublicKey{}
			key := genPubKey.KeyFromHash(stakeReq.GenPubKey)
			if err := eh.DB.MGet(ctx, key, &genPubKey); err != nil {
				eh.Logger.ErrorOnError(err, "Failed to load generated public key", []logger.Field{{"key", key}}...)
				break
			}

			cmpGenPubKeyHex, err := genPubKey.GenPubKey.CompressPubKeyHex()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to compress generated public key")
				break
			}

			cChainAddr, err := genPubKey.GenPubKey.CChainAddress()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to get C-Chain address")
				break
			}

			eh.addrLock.Lock()
			eh.cChainAddress = &cChainAddr
			eh.addrLock.Unlock()

			group := storage.Group{
				ID: genPubKey.GroupId,
			}
			if err := eh.DB.LoadModel(ctx, &group); err != nil {
				eh.Logger.ErrorOnError(err, "Failed to load group", []logger.Field{{"key", group.ID}}...)
				break
			}

			cmpPartiPubKeys, err := group.Group.CompressPubKeyHexs()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to compress participant public keys")
				break
			}

			nonce := eh.Noncer.GetNonce(stakeReq.ReqNo) // todo: how should nonce base adjust in case of validation errors among all participants?
			taskID := stakeTaskIDPrefix + evt.Raw.TxHash.Hex()

			signRequester := &SignRequester{
				SignDoner: eh.SignDoner,
				SignRequestArgs: SignRequestArgs{
					TaskID:                 taskID,
					CompressedPartiPubKeys: cmpPartiPubKeys,
					CompressedGenPubKeyHex: cmpGenPubKeyHex,
				},
			}

			eh.Logger.Debug("Nonce fetched", []logger.Field{
				{"requestID", stakeReq.ReqNo},
				{"nonce", nonce},
				{"taskID", evt.Raw.TxHash.Hex()}}...)

			taskCreator := StakeTaskCreator{
				TaskID:         taskID,
				StakeRequest:   &stakeReq,
				NetworkContext: eh.NetworkContext,
				GenPubKey:      genPubKey.GenPubKey,
				Nonce:          nonce,
			}
			stakeTask, err := taskCreator.CreateStakeTask()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to create stake task")
				break
			}

			stw := &StakeTaskWrapper{
				CChainIssueClient: eh.CChainIssueClient,
				Logger:            eh.Logger,
				PChainIssueClient: eh.PChainIssueClient,
				SignRequester:     signRequester,
				StakeTask:         stakeTask,
			}

			amount := big.NewInt(int64(stw.DelegateAmt))

			// params validation before request tx sign
			if err := eh.checkBalance(ctx, cChainAddr, amount); err != nil {
				eh.Logger.ErrorOnError(err, "Failed to check balance before request tx sign", []logger.Field{{"insufficientFundsStakeTask", stw.StakeTask}}...)
				break
			}
			if err := eh.checkStarTime(stakeReq.StartTime); err != nil {
				eh.Logger.ErrorOnError(err, "Failed to check stake start time before request tx sign", []logger.Field{{"startTimeExpiredStakeTask", stw.StakeTask}}...)
				break
			}

			eh.signStakeTxWs.AddTask(ctx, &work.Task{
				Args: []interface{}{stw, &stakeReq},
				Ctx:  ctx,
				WorkFns: []work.WorkFn{func(ctx context.Context, args interface{}) {
					stw := args.([]interface{})[0].(*StakeTaskWrapper)
					stakeReq := args.([]interface{})[1].(*storage.StakeRequest)
					if err := stw.SignTx(ctx); err != nil {
						eh.Logger.ErrorOnError(err, "Failed to sign Tx", []logger.Field{{"errSignStakeTask", stw.StakeTask}}...)
						return
					}

					// params validation after tx signed, check this because signing consume gas and time
					amount := big.NewInt(int64(stw.DelegateAmt))
					if err := eh.checkBalance(ctx, stw.CChainAddress, amount); err != nil {
						eh.Logger.DebugOnError(err, "Failed to check balance after tx signed", []logger.Field{{"insufficientFundsStakeTask", stw.StakeTask}}...)
						return
					}
					if err := eh.checkStarTime(stakeReq.StartTime); err != nil {
						eh.Logger.ErrorOnError(err, "Failed to check stake start time after tx signed", []logger.Field{{"startTimeExpiredStakeTask", stw.StakeTask}}...)
						return
					}
					issueTx := &issueTx{stw, stakeReq}
					select {
					case <-ctx.Done():
						return
					case eh.issueTxChan <- issueTx:
					}
				}},
			})
		}
	}
}

func (eh *StakeRequestStarted) issueTx(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case issueTx := <-eh.issueTxChan:
			// Cache tx for later issue
			eh.issueTxCacheLock.Lock()
			eh.issueTxCache[issueTx.Nonce] = issueTx
			eh.issueTxCacheLock.Unlock()

			// Sync nonce
			if err := eh.syncNonce(ctx); err != nil {
				eh.Logger.ErrorOnError(err, "Failed to sync nonce")
				break
			}

			// Continuous issue tx
			for i := int(eh.nextNonce); i < int(eh.nextNonce)+len(eh.issueTxCache); i++ {
				eh.issueTxCacheLock.RLock()
				issueTx, ok := eh.issueTxCache[uint64(i)]
				eh.issueTxCacheLock.RUnlock()
				if !ok {
					break
				}
				eh.doIssueTx(ctx, issueTx.StakeTaskWrapper)
				eh.nextNonce++
			}

			eh.issueTxCacheLock.Lock()
			for nonce, _ := range eh.issueTxCache {
				if nonce < eh.nextNonce {
					delete(eh.issueTxCache, nonce)
				}
			}
			eh.issueTxCacheLock.Unlock()
		}
	}
}

func (eh *StakeRequestStarted) doIssueTx(ctx context.Context, stw *StakeTaskWrapper) error {
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

	newEvt := events.StakeTaskDone{
		TaskID: common.HexToHash(strings.TrimPrefix(stakeTask.TaskID, stakeTaskIDPrefix)),

		ExportTxID:       ids[0],
		ImportTxID:       ids[1],
		AddDelegatorTxID: ids[2],

		RequestID:   stakeTask.RequestNo,
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
	//eh.Publisher.Publish(ctx, dispatcher.NewEvtObj(&newEvt, nil))

	eh.doneStakeTasks++
	prom.StakeTaskDone.Inc()

	eh.Logger.Info("Stake task DONE", []logger.Field{{"stakeTaskDoneEvent", newEvt}}...)
	return nil
}

func (eh *StakeRequestStarted) checkIssueTxCache(ctx context.Context) {
	issueT := time.NewTicker(time.Second * 60)
	statsT := time.NewTicker(time.Minute * 5)
	defer issueT.Stop()
	defer statsT.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-issueT.C:
			if len(eh.issueTxCache) != 0 {
				var issueTxs []*issueTx
				eh.issueTxCacheLock.RLock()
				for _, issueTx := range eh.issueTxCache {
					issueTxs = append(issueTxs, issueTx)
				}
				eh.issueTxCacheLock.RUnlock()
				for _, issueTx := range issueTxs {
					select {
					case <-ctx.Done():
						return
					case eh.issueTxChan <- issueTx:
					}
				}
			}
		case <-statsT.C:
			var nonces []int
			eh.issueTxCacheLock.RLock()
			for nonce, _ := range eh.issueTxCache {
				nonces = append(nonces, int(nonce))
			}
			eh.issueTxCacheLock.RUnlock()
			sort.Ints(nonces)

			if len(nonces) == 0 {
				break
			}
			eh.Logger.Debug("Stake tasks stats", []logger.Field{
				{"cachedStakeTasks", len(eh.issueTxCache)},
				{"doneStakeTasks", eh.doneStakeTasks},
				{"errIssueStakeTasks", eh.errIssueStakeTasks},
				{"cachedNonces", nonces}}...)
		}
	}
}

func (eh *StakeRequestStarted) syncNonce(ctx context.Context) error {
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

func (eh *StakeRequestStarted) checkBalance(ctx context.Context, addr common.Address, stakeAmt *big.Int) error {
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

func (eh *StakeRequestStarted) checkStarTime(startTime int64) error {
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
