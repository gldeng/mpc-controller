package stake

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

const (
	StatusStarted Status = iota
	StatusBuilt

	StatusExportTxSigningPosted
	StatusExportTxSigningDone
	StatusExportTxIssued
	StatusExportTxApproved
	StatusExportTxFailed

	StatusImportTxSigningPosted
	StatusImportTxSigningDone
	StatusImportTxIssued
	StatusImportTxApproved
	StatusImportTxFailed
)

type Status int

type StakeTransferTask struct {
	Ctx    context.Context
	Logger logger.Logger

	NonceGiver noncer.Noncer
	Network    chain.NetworkContext

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeAtomicTransferTask]

	Joined *events.RequestStarted

	txs      *Txs
	signReqs []*core.SignRequest

	exportIssueTx   *txissuer.Tx
	exportTxSignRes *core.Result

	importIssueTx   *txissuer.Tx
	importTxSignRes *core.Result

	status Status
}

func (t *StakeTransferTask) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

// todo: function extraction
// todo: add task failure log

func (t *StakeTransferTask) do() bool {
	switch t.status {
	case StatusStarted:
		err := t.buildTask()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to build")
			return false
		}
		t.status = StatusBuilt
	case StatusBuilt:
		err := t.MpcClient.Sign(t.Ctx, t.signReqs[0])
		t.Logger.ErrorOnError(err, "StakeTransferTask failed to post ExportTx signing request")
		if err == nil {
			t.status = StatusExportTxSigningPosted
		}
	case StatusExportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.signReqs[0].ReqID)
		t.Logger.ErrorOnError(err, "StakeTransferTask failed to check ExportTx signing result")

		if res.Status != core.StatusDone {
			if strings.Contains(string(res.Status), "ERROR") {
				t.Logger.ErrorOnError(errors.New(string(res.Status)), "StakeTransferTask failed to sign ExportTx")
				return false
			}
			t.Logger.Debug("StakeTransferTask hasn't finished ExportTx signing")
			return true
		}
		t.status = StatusExportTxSigningDone
		t.exportTxSignRes = res
	case StatusExportTxSigningDone:
		sig := new(events.Signature).FromHex(t.exportTxSignRes.Result)
		err := t.txs.SetExportTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to set ExportTx signature")
			return false
		}

		signedBytes, err := t.txs.SignedExportTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to get signed ExportTx bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.signReqs[0].ReqID,
			TxID:  t.txs.ExportTxID(),
			Chain: txissuer.ChainC,
			Bytes: signedBytes,
		}
		t.exportIssueTx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.exportIssueTx)
		if err == nil {
			t.status = StatusExportTxIssued
		}
	case StatusExportTxIssued:
		err := t.TxIssuer.TrackTx(t.Ctx, t.exportIssueTx)
		if err == nil && t.exportIssueTx.Status == txissuer.StatusFailed {
			t.status = StatusExportTxFailed
		}

		if err == nil && t.exportIssueTx.Status == txissuer.StatusApproved {
			t.status = StatusExportTxApproved
		}
	case StatusExportTxFailed:
		fallthrough
	case StatusExportTxApproved:
		signReq, err := t.buildImportTxSignReq(t.txs)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to build ImportTx signing request")
			return false
		}

		t.signReqs[1] = signReq

		err = t.MpcClient.Sign(t.Ctx, t.signReqs[1])
		t.Logger.ErrorOnError(err, "StakeTransferTask failed to post ImportTx signing request")
		if err == nil {
			t.status = StatusImportTxSigningPosted
		}
	case StatusImportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.signReqs[1].ReqID)
		t.Logger.ErrorOnError(err, "StakeTransferTask failed to check ImportTx signing result")

		if res.Status != core.StatusDone {
			if strings.Contains(string(res.Status), "ERROR") {
				t.Logger.ErrorOnError(errors.New(string(res.Status)), "StakeTransferTask failed to sign ImportTx")
				return false
			}
			t.Logger.Debug("StakeTransferTask hasn't finished ImportTx signing")
			return true
		}
		t.status = StatusImportTxSigningDone
		t.importTxSignRes = res
	case StatusImportTxSigningDone:
		sig := new(events.Signature).FromHex(t.importTxSignRes.Result)
		err := t.txs.SetImportTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to set ImportTx signature")
			return false
		}

		signedBytes, err := t.txs.SignedImportTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to get signed ImportTx bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.signReqs[1].ReqID,
			TxID:  t.txs.ImportTxID(),
			Chain: txissuer.ChainP,
			Bytes: signedBytes,
		}
		t.importIssueTx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.importIssueTx)
		if err == nil {
			t.status = StatusImportTxIssued
		}
	case StatusImportTxIssued:
		err := t.TxIssuer.TrackTx(t.Ctx, t.importIssueTx)
		if err == nil && t.importIssueTx.Status == txissuer.StatusFailed {
			t.status = StatusImportTxFailed
		}

		if err == nil && t.importIssueTx.Status == txissuer.StatusApproved {
			t.status = StatusImportTxApproved
		}
	case StatusImportTxFailed:
		fallthrough
	case StatusImportTxApproved:
		utxos, _ := t.txs.SingedImportTxUTXOs()

		evt := events.StakeAtomicTransferTask{
			StakeTaskBasic: events.StakeTaskBasic{
				ReqNo:   t.txs.ReqNo,
				Nonce:   t.txs.Nonce,
				ReqHash: t.txs.ReqHash,

				DelegateAmt: t.txs.DelegateAmt,
				StartTime:   t.txs.StartTime,
				EndTime:     t.txs.EndTime,
				NodeID:      t.txs.NodeID,

				StakePubKey:   t.Joined.CompressedGenPubKeyHex,
				CChainAddress: t.txs.CChainAddress,
				PChainAddress: t.txs.PChainAddress,

				JoinedPubKeys: t.Joined.CompressedPartiPubKeys,
			},

			ExportTxID: t.exportIssueTx.TxID,
			ImportTxID: t.importIssueTx.TxID,

			UTXOsToStake: utxos,
		}

		t.Dispatcher.Dispatch(&evt)
		t.Logger.Info("StakeTransferTask finished", []logger.Field{{"StakeAtomicTransferTask", evt}}...)
		return false
	}
	return true
}

// Build task

func (t *StakeTransferTask) buildTask() error {
	stakeReq := t.Joined.JoinedReq.Args.(*storage.StakeRequest)
	txs, err := t.buildTxs(stakeReq, *t.Joined.ReqHash)
	if err != nil {
		return errors.Wrapf(err, "failed to build txs")
	}

	signReq, err := t.buildExportTxSignReq(txs)
	if err != nil {
		return errors.Wrapf(err, "failed to build ExportTx signing request")
	}

	t.txs = txs
	t.signReqs = make([]*core.SignRequest, 2)
	t.signReqs[0] = signReq
	return nil
}

func (t *StakeTransferTask) buildExportTxSignReq(txs *Txs) (*core.SignRequest, error) {
	exportTxHash, err := txs.ExportTxHash()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ExportTx hash")
	}

	exportTxSignReq := core.SignRequest{
		ReqID:                  string(events.ReqIDPrefixStakeExport) + fmt.Sprintf("%v", txs.ReqNo) + "-" + txs.ReqHash,
		CompressedGenPubKeyHex: t.Joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: t.Joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(exportTxHash),
	}

	return &exportTxSignReq, nil
}

func (t *StakeTransferTask) buildImportTxSignReq(txs *Txs) (*core.SignRequest, error) {
	importTxHash, err := txs.ImportTxHash()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ImportTx hash")
	}

	importTxSignReq := core.SignRequest{
		ReqID:                  string(events.ReqIDPrefixStakeImport) + fmt.Sprintf("%v", txs.ReqNo) + "-" + txs.ReqHash,
		CompressedGenPubKeyHex: t.Joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: t.Joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(importTxHash),
	}

	return &importTxSignReq, nil
}

func (t *StakeTransferTask) buildTxs(stakeReq *storage.StakeRequest, reqHash storage.RequestHash) (*Txs, error) {
	nodeID, _ := ids.ShortFromPrefixedString(stakeReq.NodeID, ids.NodeIDPrefix)
	amountBig := new(big.Int)
	amount, _ := amountBig.SetString(stakeReq.Amount, 10)

	startTime := big.NewInt(stakeReq.StartTime)
	endTIme := big.NewInt(stakeReq.EndTime)

	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !startTime.IsUint64() || !endTIme.IsUint64() {
		return nil, errors.New(ErrMsgInvalidUint64)
	}

	cChainAddr, _ := stakeReq.GenPubKey.CChainAddress()
	pChainAddr, _ := stakeReq.GenPubKey.PChainAddress()

	st := Txs{
		ReqNo:         stakeReq.ReqNo,
		Nonce:         t.NonceGiver.GetNonce(stakeReq.ReqNo),
		ReqHash:       reqHash.String(),
		DelegateAmt:   nAVAXAmount.Uint64(),
		StartTime:     startTime.Uint64(),
		EndTime:       endTIme.Uint64(),
		CChainAddress: cChainAddr,
		PChainAddress: pChainAddr,
		NodeID:        ids.NodeID(nodeID),

		BaseFeeGwei: cchain.BaseFeeGwei,
		NetworkID:   t.Network.NetworkID(),
		CChainID:    t.Network.CChainID(),
		Asset:       t.Network.Asset(),
		ImportFee:   t.Network.ImportFee(),
	}

	return &st, nil
}
