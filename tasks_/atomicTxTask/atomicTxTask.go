package atomicTxTask

import (
	"context"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
)

const (
	StatusCreated Status = iota

	StatusExportTxSigningPosted
	StatusExportTxSigningDone
	StatusExportTxIssued
	StatusExportTxApproved

	StatusImportTxSigningPosted
	StatusImportTxSigningDone
	StatusImportTxIssued
	StatusImportTxApproved
)

type Status int

type AtomicTxTask struct {
	Status Status

	ExportTx        *txissuer.Tx
	ExportTxSignReq *core.SignRequest
	exportTxSignRes *core.Result

	ImportTx        *txissuer.Tx
	ImportTxSignReq *core.SignRequest
	importTxSignRes *core.Result

	Ctx    context.Context
	Logger logger.Logger

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Pool       pool.WorkerPool
	Dispatcher dispatcher.EventDispatcher
}

func (t *AtomicTxTask) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

func (t *AtomicTxTask) do() bool {
	switch t.Status {
	case StatusCreated:
		err := t.MpcClient.Sign(t.Ctx, t.ExportTxSignReq)
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusExportTxSigningPosted
		}
		return true
	case StatusExportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.ExportTxSignReq.ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.Status = StatusExportTxSigningDone
		t.exportTxSignRes = res
		return true
	case StatusExportTxSigningDone: // todo: set signature
		err := t.TxIssuer.IssueTx(t.Ctx, t.ExportTx)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.Status = StatusExportTxIssued
		}
		return true
	case StatusExportTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.ExportTx)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.Status = StatusExportTxApproved
		}
		return true
	case StatusExportTxApproved:
		err := t.MpcClient.Sign(t.Ctx, t.ImportTxSignReq)
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusImportTxSigningPosted
		}
		return true
	case StatusImportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.ImportTxSignReq.ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.Status = StatusImportTxSigningDone
		t.importTxSignRes = res
		return true
	case StatusImportTxSigningDone: // todo: set signature
		err := t.TxIssuer.IssueTx(t.Ctx, t.ImportTx)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.Status = StatusImportTxIssued
		}
		return true
	case StatusImportTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.ImportTx)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.Status = StatusImportTxApproved
		}
		t.Dispatcher.Dispatch(t.ImportTx) // todo: improve data sharing strategy
	}
	return false
}
