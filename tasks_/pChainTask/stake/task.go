package stake

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/utils/bytes"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
)

const (
	StatusStarted Status = iota
	StatusBuilt

	StatusTxSigningPosted
	StatusTxSigningDone
	StatusTxIssued
	StatusImportTxCommitted
)

type Status int

type Task struct {
	Ctx    context.Context
	Logger logger.Logger

	Network chain.NetworkContext

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeAddDelegatorTaskDone]

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Atomic *events.StakeAtomicTaskDone

	tx        *AddDelegatorTx
	txSignReq *core.SignRequest
	txSignRes *core.Result

	tx_ *txissuer.Tx

	status Status
}

func (t *Task) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

func (t *Task) do() bool {
	switch t.status {
	case StatusStarted:
		err := t.buildTask()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to build task")
			return false
		}
		t.status = StatusBuilt
	case StatusBuilt:
		err := t.MpcClient.Sign(t.Ctx, t.txSignReq)
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.status = StatusTxSigningPosted
		}
		return true
	case StatusTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.txSignReq.ReqID)
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
		err := t.tx.SetTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to set signature")
			return false
		}

		signedBytes, err := t.tx.SignedTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.txSignReq.ReqID,
			Kind:  events.TxKindPChainAddDelegator,
			Bytes: signedBytes,
		}
		t.tx_ = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.tx_)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.status = StatusTxIssued
		}
		return true
	case StatusTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.tx_)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.status = StatusImportTxCommitted
			t.tx.SetTxID(t.tx_.TxID)
		}

		evt := events.StakeAddDelegatorTaskDone{
			StakeTaskBasic:   t.tx.StakeTaskBasic,
			AddDelegatorTxID: t.tx.addDelegatorTxID,
		}

		t.Dispatcher.Dispatch(&evt)
	}
	return false
}

// Build task

func (t *Task) buildTask() error {
	tx, err := t.buildTx()
	if err != nil {
		return errors.WithStack(err)
	}

	signReqs, err := t.buildSignReqs(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	t.tx = tx
	t.txSignReq = signReqs
	return nil
}

func (t *Task) buildSignReqs(tx *AddDelegatorTx) (*core.SignRequest, error) {
	txHash, err := tx.TxHash()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signReq := core.SignRequest{
		ReqID:                  string(events.ReqIDPrefixStakeAddDelegator) + fmt.Sprintf("%v", tx.ReqNo) + "-" + tx.ReqHash,
		Kind:                   events.SignKindStakeAddDelegator,
		CompressedGenPubKeyHex: tx.StakePubKey,
		CompressedPartiPubKeys: tx.JoinedPubKeys,
		Hash:                   bytes.BytesToHex(txHash),
	}

	return &signReq, nil
}

func (t *Task) buildTx() (*AddDelegatorTx, error) {
	st := AddDelegatorTx{
		StakeAtomicTaskDone: t.Atomic,
		NetworkID:           t.Network.NetworkID(),
		Asset:               t.Network.Asset(),
	}

	return &st, nil
}
