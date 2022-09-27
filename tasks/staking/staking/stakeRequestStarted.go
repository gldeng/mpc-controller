package staking

import (
	"context"
	"fmt"
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
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	stakeTaskIDPrefix = "STAKE-"
)

// Subscribe event: *events.RequestStarted

type StakeRequestStarted struct {
	BoundTransactor transactor.Transactor
	DB              storage.DB
	EthClient       chain.EthClient
	IssuerCChain    chain.CChainIssuer
	IssuerPChain    chain.PChainIssuer
	Logger          logger.Logger
	NetWorkCtx      chain.NetworkContext
	NonceGiver      noncer.Noncer
	PartiPubKey     storage.PubKey
	SignerMPC       core.SignDoner

	requestStartedChan chan *events.RequestStarted
	signStakeTxWs      *work.Workshop

	issueTxChan chan *issueTx
	once        sync.Once

	issueTxCache     map[uint64]*issueTx
	issueTxCacheLock sync.RWMutex
	issueTxContainer *issueTxContainer

	nextNonce uint64 // todo: future key-rotation influence

	doneStakeTasks     uint64
	errIssueStakeTasks uint64
}

type issueTx struct {
	*StakeTaskWrapper
	*storage.StakeRequest
}

func (eh *StakeRequestStarted) Init(ctx context.Context) {
	eh.signStakeTxWs = work.NewWorkshop(eh.Logger, "signStakeTx", time.Minute*10, 10)

	eh.requestStartedChan = make(chan *events.RequestStarted, 1024)

	eh.issueTxChan = make(chan *issueTx)
	eh.issueTxCache = make(map[uint64]*issueTx)
	eh.issueTxContainer = new(issueTxContainer)

	go eh.signTx(ctx)
	go eh.issueTx(ctx)
}

func (eh *StakeRequestStarted) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.RequestStarted:
		reqHash := (storage.RequestHash)(evt.RequestHash)
		if !reqHash.IsTaskType(storage.TaskTypStake) {
			break
		}
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
			reqHash := (storage.RequestHash)(evt.RequestHash)
			stakeReq := storage.StakeRequest{}
			joinReq := storage.JoinRequest{
				ReqHash: reqHash,
				Args:    &stakeReq,
			}
			if err := eh.DB.LoadModel(ctx, &joinReq); err != nil {
				eh.Logger.DebugOnError(err, "No JoinRequest load for stake", []logger.Field{{"reqHash", reqHash.String()}}...)
				break
			}
			if !joinReq.PartiId.Joined(evt.ParticipantIndices) {
				eh.Logger.Debug("Not joined stake request", []logger.Field{{"reqHash", evt.RequestHash}}...)
				break
			}

			cmpGenPubKeyHex, err := stakeReq.GenPubKey.CompressPubKeyHex()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to compress generated public key")
				break
			}

			cChainAddr, err := stakeReq.GenPubKey.CChainAddress()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to get C-Chain address")
				break
			}

			group := storage.Group{
				ID: stakeReq.GroupId,
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

			var joinedCmpPartiPubKeys []string
			indices := (*storage.Indices)(evt.ParticipantIndices).Indices()
			for _, index := range indices {
				joinedCmpPartiPubKeys = append(joinedCmpPartiPubKeys, cmpPartiPubKeys[index-1])
			}

			nonce := eh.NonceGiver.GetNonce(stakeReq.ReqNo) // todo: how should nonce base adjust in case of validation errors among all participants?
			taskID := stakeTaskIDPrefix + reqHash.String()

			signRequester := &SignRequester{
				SignDoner: eh.SignerMPC,
				SignRequestArgs: SignRequestArgs{
					TaskID:                 taskID,
					CompressedPartiPubKeys: joinedCmpPartiPubKeys,
					CompressedGenPubKeyHex: cmpGenPubKeyHex,
				},
			}

			eh.Logger.Debug("Nonce fetched", []logger.Field{
				{"requestID", stakeReq.ReqNo},
				{"nonce", nonce},
				{"taskID", reqHash.String()}}...)

			taskCreator := StakeTaskCreator{
				TaskID:         taskID,
				StakeRequest:   &stakeReq,
				NetworkContext: eh.NetWorkCtx,
				Nonce:          nonce,
			}
			stakeTask, err := taskCreator.CreateStakeTask()
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to create stake task")
				break
			}

			stw := &StakeTaskWrapper{
				ReqHash:           reqHash,
				CChainIssueClient: eh.IssuerCChain,
				Logger:            eh.Logger,
				PChainIssueClient: eh.IssuerPChain,
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

			err = eh.signStakeTxWs.AddTask(ctx, &work.Task{
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
					eh.issueTxContainer.AddSort(issueTx)
					reqHash := stakeReq.ReqHash()
					eh.Logger.Info("Stake sign task done", []logger.Field{
						{"reqNo", stakeReq.ReqNo},
						{"reqHash", reqHash.String()},
						{"taskId", taskID}}...)
				}},
			})
			eh.Logger.ErrorOnError(err, "Failed to add sign stake task", []logger.Field{
				{"reqNo", stakeReq.ReqNo},
				{"reqHash", reqHash.String()},
				{"taskId", taskID}}...)
			eh.Logger.InfoNilError(err, "Stake sign task added", []logger.Field{
				{"reqNo", stakeReq.ReqNo},
				{"reqHash", reqHash.String()},
				{"taskId", taskID}}...)
		}
	}
}

func (eh *StakeRequestStarted) issueTx(ctx context.Context) {
	issueT := time.NewTicker(time.Second)
	defer issueT.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-issueT.C:
			if eh.issueTxContainer.IsEmpty() {
				break
			}
			addressNonce, err := eh.EthClient.NonceAt(ctx, eh.issueTxContainer.Address(), nil)
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to query nonce")
				break
			}

			txs := eh.issueTxContainer.ContinuousTxs(addressNonce)
			if len(txs) == 0 {
				break
			}
			for _, tx := range txs {
				eh.doIssueTx(ctx, tx.StakeTaskWrapper)
			}

			eh.issueTxContainer.TrimLeft(txs[len(txs)-1].Nonce)

			nonces := eh.issueTxContainer.Nonces()
			eh.Logger.Debug("Stake tasks stats", []logger.Field{
				{"cachedStakeTasks", len(nonces)},
				{"doneStakeTasks", eh.doneStakeTasks},
				{"errIssueStakeTasks", eh.errIssueStakeTasks},
				{"cachedNonces", nonces}}...)
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
			eh.Logger.DebugOnError(err, "Stake task not done for "+chain.ErrMsgInsufficientFunds, []logger.Field{{"errIssueStakeTask", stakeTask}}...)
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

	eh.doneStakeTasks++
	prom.StakeTaskDone.Inc()
	eh.Logger.Info("Stake task done", []logger.Field{{"stakeTaskDone", fmt.Sprintf("reqHash:%v, stakeTask:%+v", stw.ReqHash.String(), newEvt)}}...)
	return nil
}

func (eh *StakeRequestStarted) checkBalance(ctx context.Context, addr common.Address, stakeAmt *big.Int) error {
	bl, err := eh.EthClient.BalanceAt(ctx, addr, nil)
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
