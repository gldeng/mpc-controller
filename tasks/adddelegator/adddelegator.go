package addDelegator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
)

// todo: use ErrMsg

var (
	_ core.Task = (*AddDelegator)(nil)
)

type AddDelegator struct {
	FlowId string
	Quorum types.QuorumInfo
	Param  *StakeParam
	TxID   *ids.ID

	tx      *AddDelegatorTx
	signReq *types.SignRequest

	status       Status
	failed       bool
	StartTime    time.Time
	LastStepTime time.Time
}

func NewAddDelegator(flowId string, quorum types.QuorumInfo, param *StakeParam) (*AddDelegator, error) {
	return &AddDelegator{
		FlowId:       flowId,
		Quorum:       quorum,
		Param:        param,
		TxID:         nil,
		tx:           nil,
		signReq:      nil,
		status:       StatusInit,
		failed:       false,
		StartTime:    time.Now(),
		LastStepTime: time.Now(),
	}, nil
}

func (t *AddDelegator) GetId() string {
	return fmt.Sprintf("%v-addDelegator", t.FlowId)
}

func (t *AddDelegator) FailedPermanently() bool {
	return t.failed
}

func (t *AddDelegator) IsSequential() bool {
	return false
}

func (t *AddDelegator) IsDone() bool {
	return t.status == StatusDone
}

func (t *AddDelegator) Next(ctx core.TaskContext) ([]core.Task, error) {
	if time.Now().Sub(t.LastStepTime) < 2*time.Second { // Min delay between steps
		return nil, nil
	}
	if time.Now().Sub(t.StartTime) >= 30*time.Minute {
		return nil, errors.New(ErrMsgTimedOut)
	}
	defer func() {
		t.LastStepTime = time.Now()
	}()
	return t.run(ctx)
}

func (t *AddDelegator) run(ctx core.TaskContext) ([]core.Task, error) {
	switch t.status {
	case StatusInit:
		err := t.buildAndSignTx(ctx)
		if err != nil {
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToBuildAndSignTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToBuildAndSignTx)
		}
		t.status = StatusSignReqSent
	case StatusSignReqSent:
		err := t.getSignatureAndSendTx(ctx)
		if err != nil {
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToGetSignatureAndSendTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToGetSignatureAndSendTx)
		}

		if t.TxID != nil {
			ctx.GetLogger().Debugf("id %v AddDelegatorTx ID is %v", t.FlowId, t.tx.ID().String())
			t.status = StatusTxSent
		}
	case StatusTxSent:
		status, err := ctx.CheckPChainTx(t.tx.ID())
		ctx.GetLogger().Debugf("id %v AddDelegatorTx status is %v", t.FlowId, status)
		if err != nil {
			return nil, t.failIfErrorf(err, ErrMsgFailedToCheckStatus)
		}
		if !core.IsPending(status) {
			prom.AddDelegatorTxCommitted.Inc()
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
		return errors.Wrapf(err, "failed to build AddDelegator task")
	}
	err = ctx.GetMpcClient().Sign(context.Background(), t.signReq)
	if err != nil {
		return errors.Wrapf(err, "failed to send AddDelegator signing request")
	}
	ctx.GetLogger().Debugf("sent signing AddDelegatorTx request, requestID:%v", t.signReq.ReqID)
	return nil
}

func (t *AddDelegator) getSignatureAndSendTx(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().Result(context.Background(), t.signReq.ReqID)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.Status != types.StatusDone {
		status := strings.ToLower(string(res.Status))
		if strings.Contains(status, "error") || strings.Contains(status, "err") {
			return t.failIfErrorf(err, "failed to sign AddDelegatorTx, status:%v", status)
		}
		ctx.GetLogger().Debugf("signing AddDelegatorTx not done, requestID:%v, status:%v", t.signReq.ReqID, status) // TODO: timeout and quit
		return nil
	}

	sig := new(types.Signature).FromHex(res.Result)
	err = t.tx.SetTxSig(*sig)
	if err != nil {
		return t.failIfErrorf(err, "failed to set signature")
	}

	signedBytes, err := t.tx.SignedTxBytes()
	if err != nil {
		return t.failIfErrorf(err, "failed to get signed AddDelegatorTx bytes")
	}

	txID := t.tx.ID()
	t.TxID = &txID

	_, err = ctx.IssuePChainTx(signedBytes) // TODO: check tx status before issuing, which may has been committed by other mpc-controller?
	if err != nil {
		// TODO: Better handling, if another participant already sent the tx, we can't send again. So err is not exception, but a normal outcome.
		//return t.failIfErrorf(err, ErrMsgFailedToIssueTx)
		return nil
	}
	prom.AddDelegatorTxIssued.Inc()
	return nil
}

func (t *AddDelegator) buildTask(ctx core.TaskContext) error {
	tx, err := NewAddDelegatorTx(t.Param, t.Quorum, ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to build AddDelegatorTx")
	}

	txHash, err := tx.TxHash()
	if err != nil {
		return errors.Wrapf(err, "failed to get AddDelegatorTx hash")
	}

	signReqs, err := t.buildSignReqs(t.GetId(), txHash)
	if err != nil {
		return errors.Wrapf(err, "failed to build AddDelegatorTx sign request")
	}

	t.tx = tx
	t.signReq = signReqs
	return nil
}

func (t *AddDelegator) buildSignReqs(id string, hash []byte) (*types.SignRequest, error) {
	partiPubKeys, genPubKey, err := t.Quorum.CompressKeys()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compress public keys")
	}

	signReq := types.SignRequest{
		ReqID:                  id,
		CompressedGenPubKeyHex: genPubKey,
		CompressedPartiPubKeys: partiPubKeys,
		Hash:                   bytes.BytesToHex(hash),
	}

	return &signReq, nil
}

func (t *AddDelegator) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}
