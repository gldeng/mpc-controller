package addDelegator

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"strings"
)

// todo: use ErrMsg

var (
	_ core.Task = (*AddDelegator)(nil)
)

type AddDelegator struct {
	Id             string
	Request        *Request
	Quorum         types.QuorumInfo
	SignedImportTx *txs.Tx

	tx      *AddDelegatorTx
	signReq *core.SignRequest

	status Status
	failed bool
}

func NewAddDelegator(req *Request, id string, quorum types.QuorumInfo, signedImportTx *txs.Tx) (*AddDelegator, error) {
	return &AddDelegator{
		Id:             id,
		Request:        req,
		Quorum:         quorum,
		SignedImportTx: signedImportTx,
	}, nil
}

func (t *AddDelegator) GetId() string {
	return fmt.Sprintf("AddDelegator(%v)", t.Id)
}

func (t *AddDelegator) FailedPermanently() bool {
	return t.failed
}

func (t *AddDelegator) RequiresNonce() bool {
	return false
}

func (t *AddDelegator) IsDone() bool {
	return t.status == StatusDone
}

func (t *AddDelegator) Next(ctx core.TaskContext) ([]core.Task, error) {
	switch t.status {
	case StatusInit:
		err := t.buildAndSignTx(ctx)
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, ErrMsgFailedToBuildAndSignTx)
			return nil, err
		}
		t.status = StatusSignReqSent
	case StatusSignReqSent:
		err := t.getSignatureAndSendTx(ctx)
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, ErrMsgFailedToGetSignatureAndSendTx)
			return nil, err
		}

		ctx.GetLogger().Debug(fmt.Sprintf("id %v AddDelegatorTx ID is %v", t.Id, t.tx.ID().String()))
		t.status = StatusTxSent
	case StatusTxSent:
		status, err := ctx.CheckPChainTx(t.tx.ID())
		ctx.GetLogger().Debug(fmt.Sprintf("id %v AddDelegatorTx status is %v", t.Id, status))
		if err != nil {
			return nil, t.failIfError(err, ErrMsgFailedToCheckStatus)
		}
		if !core.IsPending(status) {
			t.status = StatusDone
			return nil, nil
		}
	}
	return nil, nil
}

// Build task

func (t *AddDelegator) buildAndSignTx(ctx core.TaskContext) error {
	err := t.buildTask(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to build task")
	}
	err = ctx.GetMpcClient().Sign(context.Background(), t.signReq)
	if err != nil {
		return errors.Wrapf(err, "failed to send signing request")
	}

	return nil
}

func (t *AddDelegator) getSignatureAndSendTx(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().Result(context.Background(), t.signReq.ReqID)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.Status != core.StatusDone {
		if strings.Contains(string(res.Status), "ERROR") {
			return t.failIfError(err, "failed to sign tx")
		}
		ctx.GetLogger().Debug(DebugMsgSignRequestNotDone)
		return nil
	}

	sig := new(events.Signature).FromHex(res.Result)
	err = t.tx.SetTxSig(*sig)
	if err != nil {
		return t.failIfError(err, "failed to set signature")
	}

	signedBytes, err := t.tx.SignedTxBytes()
	if err != nil {
		return t.failIfError(err, "failed to get signed AddDelegatorTx bytes")
	}

	_, err = ctx.IssuePChainTx(signedBytes)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToIssueTx)
	}

	return nil
}

func (t *AddDelegator) buildTask(ctx core.TaskContext) error {
	tx, err := t.buildTx(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	txHash, err := tx.TxHash()
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToCreateSignRequest)
	}

	signReqs, err := t.buildSignReqs(t.Id+"/addDelegator", txHash)
	if err != nil {
		return errors.WithStack(err)
	}

	t.tx = tx
	t.signReq = signReqs
	return nil
}

func (t *AddDelegator) buildSignReqs(id string, hash []byte) (*core.SignRequest, error) {
	var participantPks []string
	for _, pk := range t.Quorum.ParticipantPubKeys {
		participantPks = append(participantPks, hex.EncodeToString(pk))
	}

	signReq := core.SignRequest{
		ReqID:                  id,
		CompressedGenPubKeyHex: hex.EncodeToString(t.Quorum.PubKey),
		CompressedPartiPubKeys: participantPks,
		Hash:                   bytes.BytesToHex(hash),
	}

	return &signReq, nil
}

func (t *AddDelegator) buildTx(ctx core.TaskContext) (*AddDelegatorTx, error) {
	st := AddDelegatorTx{
		NetworkID:     ctx.GetNetwork().NetworkID(),
		Asset:         ctx.GetNetwork().Asset(),
		PChainAddress: t.Quorum.PChainAddress(),
		UTXOsToStake:  t.SignedImportTx.UTXOs(),
		NodeID:        t.Request.NodeID,
		StartTime:     t.Request.StartTime,
		EndTime:       t.Request.EndTime,
	}

	return &st, nil
}

func (t *AddDelegator) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.failed = true
	msg = fmt.Sprintf("[%v] %v", t.Id, msg)
	return errors.Wrap(err, msg)
}
