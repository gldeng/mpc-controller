package pChainTask

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
	StatusTxApproved
)

type Status int

type pChainTask struct {
	Status Status

	Tx        *txissuer.Tx
	TxSignReq *core.SignRequest
	TxSignRes *core.Result

	Ctx    context.Context
	Logger logger.Logger

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[any]
}

func (t *pChainTask) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

func (t *pChainTask) do() bool {
	switch t.Status {
	case StatusCreated:
		err := t.MpcClient.Sign(t.Ctx, t.TxSignReq)
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusTxSigningPosted
		}
		return true
	case StatusTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.TxSignReq.ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.Status = StatusTxSigningDone
		t.TxSignRes = res
		return true
	case StatusTxSigningDone: // todo: set signature
		err := t.TxIssuer.IssueTx(t.Ctx, t.Tx)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.Status = StatusTxIssued
		}
		return true
	case StatusTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.Tx)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.Status = StatusTxApproved
		}
		t.Dispatcher.Dispatch(t.Tx) // todo: improve data sharing strategy
	}
	return false
}
