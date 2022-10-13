package stake

import (
	"context"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	kbcevents "github.com/kubecost/events"
)

const (
	StatusCreated Status = iota

	StatusExportTxSigningPosted
	StatusExportTxSigningDone
	StatusExportTxIssued
	StatusExportTxAccepted

	StatusImportTxSigningPosted
	StatusImportTxSigningDone
	StatusImportTxIssued
	StatusImportTxCommitted
)

type Status int

type Task struct {
	Ctx    context.Context
	Logger logger.Logger

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeAtomicTaskDone]

	status Status

	Txs             *Txs
	ExportTxSignReq *core.SignRequest
	ImportTxSignReq *core.SignRequest

	exportTx        *txissuer.Tx
	exportTxSignRes *core.Result

	importTx        *txissuer.Tx
	importTxSignRes *core.Result
}

func (t *Task) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

func (t *Task) do() bool {
	switch t.status {
	case StatusCreated:
		err := t.MpcClient.Sign(t.Ctx, t.ExportTxSignReq)
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.status = StatusExportTxSigningPosted
		}
		return true
	case StatusExportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.ExportTxSignReq.ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.status = StatusExportTxSigningDone
		t.exportTxSignRes = res
		return true
	case StatusExportTxSigningDone:
		sig := new(events.Signature).FromHex(t.exportTxSignRes.Result)
		err := t.Txs.SetExportTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to set signature")
			return false
		}

		signedBytes, err := t.Txs.SignedExportTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.ExportTxSignReq.ReqID,
			Kind:  events.TxKindCChainExport,
			Bytes: signedBytes,
		}
		t.exportTx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.exportTx)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.status = StatusExportTxIssued
		}
		return true
	case StatusExportTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.exportTx)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.status = StatusExportTxAccepted
			t.Txs.SetExportTxID(t.exportTx.TxID)
		}
		return true
	case StatusExportTxAccepted:
		err := t.MpcClient.Sign(t.Ctx, t.ImportTxSignReq)
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.status = StatusImportTxSigningPosted
		}
		return true
	case StatusImportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.ImportTxSignReq.ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.status = StatusImportTxSigningDone
		t.importTxSignRes = res
		return true
	case StatusImportTxSigningDone:
		sig := new(events.Signature).FromHex(t.importTxSignRes.Result)
		err := t.Txs.SetImportTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to set signature")
			return false
		}

		signedBytes, err := t.Txs.SignedImportTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.ImportTxSignReq.ReqID,
			Kind:  events.TxKindPChainImport,
			Bytes: signedBytes,
		}
		t.importTx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.importTx)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.status = StatusImportTxIssued
		}
		return true
	case StatusImportTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.importTx)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.status = StatusImportTxCommitted
			t.Txs.SetImportTxID(t.importTx.TxID)
		}

		utxos, _ := t.Txs.SingedImportTxUTXOs()

		evt := events.StakeAtomicTaskDone{
			StakeTaskBasic: events.StakeTaskBasic{
				ReqNo:   t.Txs.ReqNo,
				Nonce:   t.Txs.Nonce,
				ReqHash: t.Txs.ReqHash,

				DelegateAmt: t.Txs.DelegateAmt,
				StartTime:   t.Txs.StartTime,
				EndTime:     t.Txs.EndTime,
				NodeID:      t.Txs.NodeID,

				PubKeyHex:     t.ExportTxSignReq.CompressedGenPubKeyHex,
				CChainAddress: t.Txs.CChainAddress,
				PChainAddress: t.Txs.PChainAddress,

				ParticipantPubKeys: t.ExportTxSignReq.CompressedPartiPubKeys,
			},

			ExportTxID: t.exportTx.TxID,
			ImportTxID: t.importTx.TxID,

			UTXOsToStake: utxos,
		}

		t.Dispatcher.Dispatch(&evt)
	}
	return false
}
