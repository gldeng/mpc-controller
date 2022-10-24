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
	"strings"
)

const (
	StatusStarted Status = iota
	StatusBuilt

	StatusTxSigningPosted
	StatusTxSigningDone
	StatusTxIssued
	StatusImportTxCommitted
	StatusImportTxFailed
)

type Status int

type StakeAddDelegatorTask struct {
	Ctx    context.Context
	Logger logger.Logger

	Network chain.NetworkContext

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeAddDelegatorTask]

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Atomic *events.StakeAtomicTransferTask

	tx        *AddDelegatorTx
	txSignReq *core.SignRequest
	txSignRes *core.Result

	issueTx *txissuer.Tx

	status Status
}

// todo: function extraction
// todo: add task failure log

func (t *StakeAddDelegatorTask) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

func (t *StakeAddDelegatorTask) do() bool {
	switch t.status {
	case StatusStarted:
		err := t.buildTask()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeAddDelegatorTask failed to build")
			return false
		}
		t.status = StatusBuilt
	case StatusBuilt:
		err := t.MpcClient.Sign(t.Ctx, t.txSignReq)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeAddDelegatorTask failed to post AddDelegatorTx signing request")
			return false
		}
		t.status = StatusTxSigningPosted
	case StatusTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.txSignReq.ReqID)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeAddDelegatorTask failed to check AddDelegatorTx signing result")
			return false
		}

		if res.Status != core.StatusDone {
			if strings.Contains(string(res.Status), "ERROR") {
				t.Logger.ErrorOnError(errors.New(string(res.Status)), "StakeAddDelegatorTask failed to sign AddDelegatorTx")
				return false
			}
			t.Logger.Debug("StakeAddDelegatorTask hasn't finished signing AddDelegatorTx")
			return true
		}
		t.status = StatusTxSigningDone
		t.txSignRes = res
	case StatusTxSigningDone:
		sig := new(events.Signature).FromHex(t.txSignRes.Result)
		err := t.tx.SetTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeAddDelegatorTask failed to set AddDelegatorTx signature")
			return false
		}

		signedBytes, err := t.tx.SignedTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeAddDelegatorTask failed to get signed AddDelegatorTx bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.txSignReq.ReqID,
			TxID:  t.tx.ID(),
			Chain: txissuer.ChainP,
			Bytes: signedBytes,
		}
		t.issueTx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.issueTx)
		t.Logger.ErrorOnError(err, "StakeAddDelegatorTask failed to issue AddDelegatorTx")
		if err == nil {
			t.status = StatusTxIssued
		}
	case StatusTxIssued:
		err := t.TxIssuer.TrackTx(t.Ctx, t.issueTx)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeAddDelegatorTask failed to track AddDelegatorTx status")
			return false
		}

		switch t.issueTx.Status {
		case txissuer.StatusFailed:
			t.status = StatusImportTxFailed
			t.Logger.Debug(fmt.Sprintf("StakeAddDelegatorTask failed because of %v", t.issueTx.Reason))
		case txissuer.StatusCommitted:
			t.status = StatusImportTxCommitted
			evt := events.StakeAddDelegatorTask{
				StakeTaskBasic:   t.tx.StakeTaskBasic,
				AddDelegatorTxID: t.tx.ID(),
			}

			t.Dispatcher.Dispatch(&evt)
			t.Logger.Info("StakeAddDelegatorTask finished", []logger.Field{{"StakeAddDelegatorTask", evt}}...)
		}
		return false
	}
	return true
}

// Build task

func (t *StakeAddDelegatorTask) buildTask() error {
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

func (t *StakeAddDelegatorTask) buildSignReqs(tx *AddDelegatorTx) (*core.SignRequest, error) {
	txHash, err := tx.TxHash()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signReq := core.SignRequest{
		ReqID:                  string(events.ReqIDPrefixStakeAddDelegator) + fmt.Sprintf("%v", tx.ReqNo) + "-" + tx.ReqHash,
		CompressedGenPubKeyHex: tx.StakePubKey,
		CompressedPartiPubKeys: tx.JoinedPubKeys,
		Hash:                   bytes.BytesToHex(txHash),
	}

	return &signReq, nil
}

func (t *StakeAddDelegatorTask) buildTx() (*AddDelegatorTx, error) {
	st := AddDelegatorTx{
		StakeAtomicTransferTask: t.Atomic,
		NetworkID:               t.Network.NetworkID(),
		Asset:                   t.Network.Asset(),
	}

	return &st, nil
}
