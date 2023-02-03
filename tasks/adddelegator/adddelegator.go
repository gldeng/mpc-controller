package addDelegator

import (
	"context"
	"fmt"
	"time"

	"github.com/avalido/mpc-controller/core/mpc"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/bytes"
	utilstime "github.com/avalido/mpc-controller/utils/time"
	"github.com/pkg/errors"
)

// todo: use ErrMsg

const (
	taskTypeAddDelegator = "addDelegator"
)

var (
	_ core.Task = (*AddDelegator)(nil)
)

type AddDelegator struct {
	FlowId   core.FlowId
	TaskType string
	Quorum   types.QuorumInfo
	Param    *StakeParam
	TxID     *ids.ID

	tx      *AddDelegatorTx
	signReq *mpc.SignRequest

	status       Status
	failed       bool
	StartTime    *time.Time
	LastStepTime *time.Time

	issuedByOthers bool
}

func NewAddDelegator(flowId core.FlowId, quorum types.QuorumInfo, param *StakeParam) (*AddDelegator, error) {
	return &AddDelegator{
		FlowId:   flowId,
		TaskType: taskTypeAddDelegator,
		Quorum:   quorum,
		Param:    param,
		TxID:     nil,
		tx:       nil,
		signReq:  nil,
		status:   StatusInit,
		failed:   false,
	}, nil
}

func (t *AddDelegator) GetId() string {
	return fmt.Sprintf("%v-addDelegator", t.FlowId.String())
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
	if t.StartTime == nil {
		now := time.Now()
		t.StartTime = &now
		t.LastStepTime = &now
	}

	timeout := 60 * time.Minute
	interval := 2 * time.Second // Min delay between steps
	if time.Now().Sub(*t.LastStepTime) < interval {
		return nil, nil
	}
	if time.Now().Sub(*t.StartTime) >= timeout {
		prom.TaskTimeout.With(prometheus.Labels{"flow": "initialStake", "task": taskTypeAddDelegator}).Inc()
		return nil, errors.New(ErrMsgTimedOut)
	}
	defer func() {
		now := time.Now()
		t.LastStepTime = &now
	}()
	return t.run(ctx)
}

func (t *AddDelegator) run(ctx core.TaskContext) ([]core.Task, error) {
	switch t.status {
	case StatusInit:
		err := t.buildAndSignTx(ctx)
		if err != nil {
			t.logError(ctx, ErrMsgFailedToBuildAndSignTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToBuildAndSignTx)
		}
		t.status = StatusSignReqSent
	case StatusSignReqSent:
		return nil, errors.WithStack(t.getSignature(ctx))
	case StatusSignedTxReady:
		return nil, errors.WithStack(t.sendTx(ctx))
	case StatusTxSent:
		status, err := t.checkTxStatus(ctx)
		if err != nil {
			t.logError(ctx, ErrMsgCheckTxStatusFail, err, []logger.Field{{"txId", t.TxID}}...)
			return nil, errors.Wrapf(err, "%v, txId: %v", ErrMsgCheckTxStatusFail, t.TxID)
		}

		t.logDebug(ctx, "checked tx status", []logger.Field{
			{"txId", t.TxID},
			{"status", status.Code.String()},
			{"reason", status.Reason}}...)
	}
	return nil, nil
}

// Build task

func (t *AddDelegator) buildAndSignTx(ctx core.TaskContext) error {
	err := t.buildTask(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to build AddDelegator task")
	}
	_, err = ctx.GetMpcClient().Sign(context.Background(), t.signReq)
	if err != nil {
		return errors.Wrapf(err, "failed to send AddDelegator signing request")
	}
	prom.MpcSignPostedForAddDelegatorTx.Inc()
	t.logDebug(ctx, "sent signing request", []logger.Field{{"signReq", t.signReq.RequestId}}...)
	return nil
}

func (t *AddDelegator) getSignature(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().CheckResult(context.Background(), &mpc.CheckResultRequest{RequestId: t.signReq.RequestId})
	// TODO: Handle 404
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.RequestStatus == mpc.CheckResultResponse_ERROR {
		return t.failIfErrorf(err, "failed to sign AddDelegatorTx from P-Chain, status:%v", res.RequestStatus.String())
	}

	if res.RequestStatus != mpc.CheckResultResponse_DONE {
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.signReq.RequestId}, {"status", res.RequestStatus.String()}}...)
		return nil
	}

	prom.MpcSignDoneForAddDelegatorTx.Inc()
	t.logInfo(ctx, "signing done", []logger.Field{{"signReq", t.signReq.RequestId}}...)
	sig := new(types.Signature).FromHex(res.Result)
	err = t.tx.SetTxSig(*sig)
	if err != nil {
		return t.failIfErrorf(err, "failed to set signature")
	}

	_, err = t.tx.SignedTxBytes()
	if err != nil {
		return t.failIfErrorf(err, "failed to get signed bytes")
	}

	txID := t.tx.ID()
	t.TxID = &txID

	t.status = StatusSignedTxReady
	return nil
}

// sendTx sends a tx to avalanche network. Without a consensus mechanism among the participants, every partiticipant
// attempts to issue the tx. We do the following to mitigate the race condition:
//  1. delay a random duration before sending tx
//  2. check tx status on-chain in case other participants already send it, if already sent (i.e. the tx is known to
//     avalanche network already before sending tx
//  3. check tx status again after sending failed which may be caused by another participant sending the same tx
//     at the same time
func (t *AddDelegator) sendTx(ctx core.TaskContext) error {
	// Delay for random duration to reduce race condition
	utilstime.RandomDelay(5000)

	isIssued, err := t.checkIfTxIssued(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if isIssued {
		return nil
	}

	signed, _ := t.tx.SignedTxBytes()
	_, err = ctx.IssuePChainTx(signed)
	if err != nil {
		_, err := t.checkIfTxIssued(ctx)
		return errors.WithStack(err)
	}

	t.status = StatusTxSent
	prom.AddDelegatorTxIssued.Inc()
	t.logDebug(ctx, "tx issued", []logger.Field{{Key: "txId", Value: t.TxID}}...)
	return nil
}

func (t *AddDelegator) buildTask(ctx core.TaskContext) error {
	tx, err := NewAddDelegatorTx(t.Param, t.Quorum, t.FlowId, ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to build AddDelegatorTx")
	}

	txHash, err := tx.TxHash()
	if err != nil {
		return errors.Wrapf(err, "failed to get AddDelegatorTx hash")
	}

	prom.MpcTxBuilt.With(prometheus.Labels{"flow": "initialStake", "chain": "pChain", "tx": "addDelegatorTx"}).Inc()

	signReqs, err := t.buildSignReqs(t.GetId(), txHash)
	if err != nil {
		return errors.Wrapf(err, "failed to build AddDelegatorTx sign request")
	}

	t.tx = tx
	t.signReq = signReqs
	return nil
}

func (t *AddDelegator) buildSignReqs(id string, hash []byte) (*mpc.SignRequest, error) {
	partiPubKeys, genPubKey, err := t.Quorum.CompressKeys()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compress public keys")
	}

	s := &mpc.SignRequest{}
	s.RequestId = id
	s.ParticipantPublicKeys = partiPubKeys
	s.PublicKey = genPubKey
	s.Hash = bytes.BytesToHex(hash)

	return s, nil
}

func (t *AddDelegator) checkIfTxIssued(ctx core.TaskContext) (bool, error) {
	status, err := t.checkTxStatus(ctx)
	if err != nil {
		t.logError(ctx, ErrMsgCheckTxStatusFail, err, []logger.Field{{"txId", t.TxID}}...)
		return false, errors.Wrapf(err, "%v, txId: %v", ErrMsgCheckTxStatusFail, t.TxID)
	}

	t.logDebug(ctx, "checked tx status", []logger.Field{
		{"txId", t.TxID},
		{"status", status.Code.String()},
		{"reason", status.Reason}}...)

	defer func() {
		if status.Code == core.TxStatusProcessing {
			t.status = StatusTxSent
			prom.AddDelegatorTxIssued.Inc()
		}
	}()

	switch status.Code {
	case core.TxStatusCommitted:
		return true, nil
	case core.TxStatusProcessing:
		return true, nil
	default:
		return false, nil
	}
}

func (t *AddDelegator) checkTxStatus(ctx core.TaskContext) (core.Status, error) {
	status, err := ctx.CheckPChainTx(*t.TxID)
	if err != nil {
		return status, errors.WithStack(err)
	}

	defer func() {
		if status.Code == core.TxStatusCommitted {
			t.status = StatusDone
			prom.AddDelegatorTxCommitted.Inc()
		}
	}()

	switch status.Code {
	case core.TxStatusUnknown:
		return status, nil
	case core.TxStatusCommitted:
		return status, nil
	case core.TxStatusProcessing:
		return status, nil
	default:
		return status, t.failIfErrorf(errors.New(status.String()), ErrMsgTxFail)
	}
}

func (t *AddDelegator) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *AddDelegator) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *AddDelegator) logInfo(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Info(msg, allFields...)
}

func (t *AddDelegator) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
