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

	StatusTxSigningPosted
	StatusTxSigningDone
	StatusTxIssued
	StatusImportTxCommitted
)

type Status int

type Task struct {
	Ctx    context.Context
	Logger logger.Logger

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeAddDelegatorTaskDone]

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	status Status

	Tx        *AddDelegatorTx
	TxSignReq *core.SignRequest

	tx        *txissuer.Tx
	txSignRes *core.Result
}

func (t *Task) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

func (t *Task) do() bool {
	switch t.status {
	case StatusCreated:
		err := t.MpcClient.Sign(t.Ctx, t.TxSignReq)
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.status = StatusTxSigningPosted
		}
		return true
	case StatusTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.TxSignReq.ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.status = StatusTxSigningDone
		t.txSignRes = res
		return true
	case StatusTxSigningDone: // todo: set signature
		sig := new(events.Signature).FromHex(t.txSignRes.Result)
		err := t.Tx.SetTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to set signature")
			return false
		}

		signedBytes, err := t.Tx.SignedTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.TxSignReq.ReqID,
			Kind:  events.TxKindPChainAddDelegator,
			Bytes: signedBytes,
		}
		t.tx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.tx)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.status = StatusTxIssued
		}
		return true
	case StatusTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.tx)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.status = StatusImportTxCommitted
			t.Tx.SetTxID(t.tx.TxID)
		}

		evt := events.StakeAddDelegatorTaskDone{
			StakeTaskBasic:   t.Tx.StakeTaskBasic,
			AddDelegatorTxID: t.Tx.addDelegatorTxID,
		}

		t.Dispatcher.Dispatch(&evt)
	}
	return false
}
