package staking

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
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
	SignerMPC       core.Signer

	requestStartedChan chan *events.RequestStarted
	signStakeTxWs      *work.Workshop

	issueTxChan chan *issueTx
	once        sync.Once

	issueTxCache     map[uint64]*issueTx
	issueTxCacheLock sync.RWMutex
	issueTxContainer *issueTxContainer

	pendingStakeTasks *sync.Map

	nextNonce uint64 // todo: future key-rotation influence

	doneStakeTasks     uint64
	errIssueStakeTasks uint64
}

type issueTx struct {
	*StakeTaskWrapper
	*storage.StakeRequest
}

type pendingStakeTask struct {
	stakeTask *StakeTask

	exportTxSignReq     *core.SignRequest
	importTxSignReq     *core.SignRequest
	addDelegatorSignReq *core.SignRequest

	ExportTxID       ids.ID
	ImportTxID       ids.ID
	AddDelegatorTxID ids.ID
}

func (eh *StakeRequestStarted) Init(ctx context.Context) {
	eh.signStakeTxWs = work.NewWorkshop(eh.Logger, "signStakeTx", time.Minute*10, 10)

	eh.requestStartedChan = make(chan *events.RequestStarted, 1024)

	eh.issueTxChan = make(chan *issueTx)
	eh.issueTxCache = make(map[uint64]*issueTx)
	eh.issueTxContainer = new(issueTxContainer)

	go eh.issueTx(ctx)
}

func (eh *StakeRequestStarted) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.StakeRequestStarted:
		// Create stake task
		stakeReq := evt.Args.(*storage.StakeRequest)
		stakeTask, _ := eh.createStakeTask(stakeReq, evt.ReqHash)

		// Get generated key and participant keys
		cmpGenPubKeyHex, err := stakeReq.GenPubKey.CompressPubKeyHex()
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to compress public key")
			break
		}

		group := storage.Group{
			ID: stakeReq.GroupId,
		}
		if err := eh.DB.LoadModel(ctx, &group); err != nil {
			eh.Logger.ErrorOnError(err, "Failed to load group")
			break
		}

		cmpPartiPubKeys, err := group.Group.CompressPubKeyHexs()
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to compress participant public keys")
			break
		}

		var joinedCmpPartiPubKeys []string
		indices := evt.PartiIndices.Indices()
		for _, index := range indices {
			joinedCmpPartiPubKeys = append(joinedCmpPartiPubKeys, cmpPartiPubKeys[index-1])
		}

		// Sign ExportTx
		txHash, err := stakeTask.ExportTxHash()
		if err != nil {
			eh.Logger.ErrorOnError(errors.WithStack(err), "Failed to generate ExportTx hash")
			break
		}

		signReq := core.SignRequest{
			ReqID:                  string(events.SignIDPrefixStakeExport) + fmt.Sprintf("%v", stakeTask.ReqNo) + "-" + stakeTask.ReqHash,
			Kind:                   events.SignKindStakeExport,
			CompressedGenPubKeyHex: cmpGenPubKeyHex,
			CompressedPartiPubKeys: joinedCmpPartiPubKeys,
			Hash:                   bytes.BytesToHex(txHash),
		}

		err = eh.SignerMPC.Sign(ctx, &signReq)
		eh.Logger.ErrorOnError(err, "Failed to request sign of ExportTx")
		var p pendingStakeTask
		p.stakeTask = stakeTask
		eh.pendingStakeTasks.Store(signReq.ReqID, &p)
	case *events.SignDone:
		switch evt.Kind {
		case events.SignKindStakeExport:
			// Set ExportTx signature
			pVal, _ := eh.pendingStakeTasks.Load(evt.ReqID)
			p := pVal.(pendingStakeTask)
			err := p.stakeTask.SetExportTxSig(*evt.Result)
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to set ExportTx signature")
				break
			}
			// Sign ImportTx
			txHash, err := p.stakeTask.ImportTxHash()
			if err != nil {
				eh.Logger.ErrorOnError(errors.WithStack(err), "Failed to generate ImportTx hash")
				break
			}

			signReq := core.SignRequest{
				ReqID:                  string(events.SignIDPrefixStakeImport) + fmt.Sprintf("%v", p.stakeTask.ReqNo) + "-" + p.stakeTask.ReqHash,
				Kind:                   events.SignKindStakeImport,
				CompressedGenPubKeyHex: p.exportTxSignReq.CompressedGenPubKeyHex,
				CompressedPartiPubKeys: p.exportTxSignReq.CompressedPartiPubKeys,
				Hash:                   bytes.BytesToHex(txHash),
			}

			err = eh.SignerMPC.Sign(ctx, &signReq)
			eh.Logger.ErrorOnError(err, "Failed to request sign of ImportTx")
			eh.pendingStakeTasks.Store(signReq.ReqID, &p)
			eh.pendingStakeTasks.Delete(evt.ReqID)
		case events.SignKindStakeImport:
			// Set ImportTx signature
			pVal, _ := eh.pendingStakeTasks.Load(evt.ReqID)
			p := pVal.(pendingStakeTask)
			err := p.stakeTask.SetImportTxSig(*evt.Result)
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to set ImportTx signature")
				break
			}
			// Sign AddDelegatorTx
			txHash, err := p.stakeTask.AddDelegatorTxHash()
			if err != nil {
				eh.Logger.ErrorOnError(errors.WithStack(err), "Failed to generate AddDelegatorTx hash")
				break
			}

			signReq := core.SignRequest{
				ReqID:                  string(events.SignIDPrefixStakeAddDelegator) + fmt.Sprintf("%v", p.stakeTask.ReqNo) + "-" + p.stakeTask.ReqHash,
				Kind:                   events.SignKindStakeAddDelegator,
				CompressedGenPubKeyHex: p.importTxSignReq.CompressedGenPubKeyHex,
				CompressedPartiPubKeys: p.importTxSignReq.CompressedPartiPubKeys,
				Hash:                   bytes.BytesToHex(txHash),
			}

			err = eh.SignerMPC.Sign(ctx, &signReq)
			eh.Logger.ErrorOnError(err, "Failed to request sign of AddDelegatorTx")
			eh.pendingStakeTasks.Store(signReq.ReqID, &p)
			eh.pendingStakeTasks.Delete(evt.ReqID)
		case events.SignKindStakeAddDelegator:
			// Set AddDelegatorTx signature
			pVal, _ := eh.pendingStakeTasks.Load(evt.ReqID)
			p := pVal.(pendingStakeTask)
			err := p.stakeTask.SetAddDelegatorTxSig(*evt.Result)
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to set AddDelegatorTx signature")
				break
			}
			// issue ExportTx
		}
	}
}

func (eh *StakeRequestStarted) createStakeTask(stakeReq *storage.StakeRequest, reqHash storage.RequestHash) (*StakeTask, error) {
	nodeID, _ := ids.ShortFromPrefixedString(stakeReq.NodeID, ids.NodeIDPrefix)
	amountBig := new(big.Int)
	amount, _ := amountBig.SetString(stakeReq.Amount, 10)

	startTime := big.NewInt(stakeReq.StartTime)
	endTIme := big.NewInt(stakeReq.EndTime)

	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !startTime.IsUint64() || !endTIme.IsUint64() {
		return nil, errors.New("invalid uint64")
	}

	cChainAddr, _ := stakeReq.GenPubKey.CChainAddress()
	pChainAddr, _ := stakeReq.GenPubKey.PChainAddress()

	st := StakeTask{
		ReqNo:         stakeReq.ReqNo,
		Nonce:         eh.NonceGiver.GetNonce(stakeReq.ReqNo),
		ReqHash:       reqHash.String(),
		DelegateAmt:   nAVAXAmount.Uint64(),
		StartTime:     startTime.Uint64(),
		EndTime:       endTIme.Uint64(),
		CChainAddress: cChainAddr,
		PChainAddress: pChainAddr,
		NodeID:        ids.NodeID(nodeID),
		BaseFeeGwei:   cchain.BaseFeeGwei,
		Network:       eh.NetWorkCtx,
	}

	return &st, nil
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

		RequestID:   stakeTask.ReqNo,
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
