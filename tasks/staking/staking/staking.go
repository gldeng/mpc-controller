package staking

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
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
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"time"
)

type Staking struct {
	BoundTransactor transactor.Transactor
	DB              storage.DB
	EthClient       chain.EthClient
	TxIssuer        txissuer.TxIssuer
	Logger          logger.Logger
	NetWorkCtx      chain.NetworkContext
	NonceGiver      noncer.Noncer
	PartiPubKey     storage.PubKey
	SignerMPC       core.Signer

	requestStartedChan chan *events.RequestStarted

	issueTxChan chan *pendingStakeTask
	once        sync.Once

	issueTxCache     map[uint64]*pendingStakeTask
	issueTxCacheLock sync.RWMutex
	exportTxSorter   *txSorter

	pendingStakeTasks *sync.Map

	nextNonce uint64 // todo: future key-rotation influence

	doneStakeTasks     uint64
	errIssueStakeTasks uint64
}

type pendingStakeTask struct {
	stakeTask *StakeTask

	exportTxSignReq     *core.SignRequest
	importTxSignReq     *core.SignRequest
	addDelegatorSignReq *core.SignRequest

	exportTxID       ids.ID
	importTxID       ids.ID
	addDelegatorTxID ids.ID
}

func (eh *Staking) Init(ctx context.Context) {
	eh.requestStartedChan = make(chan *events.RequestStarted, 1024)

	eh.issueTxChan = make(chan *pendingStakeTask)
	eh.issueTxCache = make(map[uint64]*pendingStakeTask)
	eh.exportTxSorter = new(txSorter)

	go eh.issueSignedExportTx(ctx)
}

func (eh *Staking) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.RequestStarted:
		if evt.TaskType == storage.TaskTypStake {
			eh.OnReqStarted(ctx, evt)
		}
	case *events.SignDone:
		eh.OnSignDone(ctx, evt)
	case *events.TxApproved:
		eh.OnTxApproved(ctx, evt)
	}
}

func (eh *Staking) OnReqStarted(ctx context.Context, evt *events.RequestStarted) {
	// Create stake task
	stakeReq := evt.JoinedReq.Args.(*storage.StakeRequest)
	stakeTask, _ := eh.createStakeTask(stakeReq, *evt.ReqHash)

	// Sign ExportTx
	txHash, err := stakeTask.ExportTxHash()
	if err != nil {
		eh.Logger.ErrorOnError(errors.WithStack(err), "Failed to generate ExportTx hash")
		return
	}

	signReq := core.SignRequest{
		ReqID:                  string(events.ReqIDPrefixStakeExport) + fmt.Sprintf("%v", stakeTask.ReqNo) + "-" + stakeTask.ReqHash,
		Kind:                   events.SignKindStakeExport,
		CompressedGenPubKeyHex: evt.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: evt.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(txHash),
	}

	err = eh.SignerMPC.Sign(ctx, &signReq)
	eh.Logger.ErrorOnError(err, "Failed to request sign of ExportTx")
	var p pendingStakeTask
	p.stakeTask = stakeTask
	eh.pendingStakeTasks.Store(signReq.ReqID, &p)
}

func (eh *Staking) OnSignDone(ctx context.Context, evt *events.SignDone) {
	pVal, _ := eh.pendingStakeTasks.Load(evt.ReqID)
	p := pVal.(*pendingStakeTask)
	switch evt.Kind {
	case events.SignKindStakeExport:
		// Set ExportTx signature
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
			ReqID:                  string(events.ReqIDPrefixStakeImport) + fmt.Sprintf("%v", p.stakeTask.ReqNo) + "-" + p.stakeTask.ReqHash,
			Kind:                   events.SignKindStakeImport,
			CompressedGenPubKeyHex: p.exportTxSignReq.CompressedGenPubKeyHex,
			CompressedPartiPubKeys: p.exportTxSignReq.CompressedPartiPubKeys,
			Hash:                   bytes.BytesToHex(txHash),
		}

		err = eh.SignerMPC.Sign(ctx, &signReq)
		eh.Logger.ErrorOnError(err, "Failed to request sign of ImportTx")
		eh.pendingStakeTasks.Store(signReq.ReqID, p)
	case events.SignKindStakeImport:
		// Set ImportTx signature
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
			ReqID:                  string(events.ReqIDPrefixStakeAddDelegator) + fmt.Sprintf("%v", p.stakeTask.ReqNo) + "-" + p.stakeTask.ReqHash,
			Kind:                   events.SignKindStakeAddDelegator,
			CompressedGenPubKeyHex: p.importTxSignReq.CompressedGenPubKeyHex,
			CompressedPartiPubKeys: p.importTxSignReq.CompressedPartiPubKeys,
			Hash:                   bytes.BytesToHex(txHash),
		}

		err = eh.SignerMPC.Sign(ctx, &signReq)
		eh.Logger.ErrorOnError(err, "Failed to request sign of AddDelegatorTx")
		eh.pendingStakeTasks.Store(signReq.ReqID, p)
	case events.SignKindStakeAddDelegator:
		// Set AddDelegatorTx signature
		err := p.stakeTask.SetAddDelegatorTxSig(*evt.Result)
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to set AddDelegatorTx signature")
			break
		}
		eh.exportTxSorter.AddSort(p)
		eh.pendingStakeTasks.Store(evt.ReqID, p)
	}
}

func (eh *Staking) OnTxApproved(ctx context.Context, evt *events.TxApproved) {
	pVal, _ := eh.pendingStakeTasks.Load(evt.ReqID)
	p := pVal.(*pendingStakeTask)
	switch evt.Kind {
	case events.TxKindCChainExport:
		// Issue signed ImportTx
		importTx, err := p.stakeTask.GetSignedImportTx()
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to get signed ImportTx")
			break
		}

		tx := txissuer.Tx{
			ReqID: p.importTxSignReq.ReqID,
			Kind:  events.TxKindPChainImport,
			Bytes: importTx.Bytes(),
		}
		err = eh.TxIssuer.IssueTx(ctx, &tx)
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to issue signed ImportTx")
			break
		}
		p.exportTxID = evt.TxID
		eh.pendingStakeTasks.Store(evt.ReqID, p)
	case events.TxKindPChainImport:
		// Issue signed ImportTx
		addDelegatorTx, err := p.stakeTask.GetSignedAddDelegatorTx()
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to get signed ImportTx")
			break
		}

		tx := txissuer.Tx{
			ReqID: p.addDelegatorSignReq.ReqID,
			Kind:  events.TxKindPChainAddDelegator,
			Bytes: addDelegatorTx.Bytes(),
		}
		err = eh.TxIssuer.IssueTx(ctx, &tx)
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to issue signed AddDelegatorImportTx")
			break
		}
		p.importTxID = evt.TxID
		eh.pendingStakeTasks.Store(evt.ReqID, p)
	case events.TxKindPChainAddDelegator:
		p.addDelegatorTxID = evt.TxID
		eh.pendingStakeTasks.Store(evt.ReqID, p)
		prom.StakeTaskDone.Inc()
		std := events.StakeAddDelegatorTaskDone{
			ReqNo:   p.stakeTask.ReqNo,
			Nonce:   p.stakeTask.Nonce,
			ReqHash: p.stakeTask.ReqHash,

			DelegateAmt: p.stakeTask.DelegateAmt,
			StartTime:   p.stakeTask.StartTime,
			EndTime:     p.stakeTask.EndTime,
			NodeID:      p.stakeTask.NodeID,

			ExportTxID:       p.exportTxID,
			ImportTxID:       p.importTxID,
			AddDelegatorTxID: p.addDelegatorTxID,

			PubKeyHex:     p.exportTxSignReq.CompressedGenPubKeyHex,
			CChainAddress: p.stakeTask.CChainAddress,
			PChainAddress: p.stakeTask.PChainAddress,

			ParticipantPubKeys: p.exportTxSignReq.CompressedPartiPubKeys,
		}
		eh.Logger.Info("Stake task done", []logger.Field{{"stakeTaskDone", std}}...)
	}
	eh.pendingStakeTasks.Delete(evt.ReqID)
}

func (eh *Staking) createStakeTask(stakeReq *storage.StakeRequest, reqHash storage.RequestHash) (*StakeTask, error) {
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

func (eh *Staking) issueSignedExportTx(ctx context.Context) {
	issueT := time.NewTicker(time.Second)
	defer issueT.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-issueT.C:
			// Sync address nonce
			if eh.exportTxSorter.IsEmpty() {
				break
			}
			addressNonce, err := eh.EthClient.NonceAt(ctx, eh.exportTxSorter.Address(), nil)
			if err != nil {
				eh.Logger.ErrorOnError(err, "Failed to query nonce")
				break
			}

			ps := eh.exportTxSorter.ContinuousTxs(addressNonce)
			if len(ps) == 0 {
				break
			}
			for _, p := range ps {
				// Issue signed ExportTx
				exportTx, err := p.stakeTask.GetSignedExportTx()
				if err != nil {
					eh.Logger.ErrorOnError(err, "Failed to get signed ExportTx")
					break
				}

				tx := txissuer.Tx{
					ReqID: p.exportTxSignReq.ReqID,
					Kind:  events.TxKindCChainExport,
					Bytes: exportTx.SignedBytes(),
				}
				err = eh.TxIssuer.IssueTx(ctx, &tx)
				if err != nil {
					eh.Logger.ErrorOnError(err, "Failed to issue signed ExportTx")
					break
				}
			}

			eh.exportTxSorter.TrimLeft(ps[len(ps)-1].stakeTask.Nonce)
		}
	}
}
