package addDelegator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/avalido/mpc-controller/taskcontext"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
)

// todo: use ErrMsg

const (
	taskTypeExport = "addDelegator"
)

var (
	_ core.Task = (*AddDelegator)(nil)
)

type AddDelegator struct {
	FlowId   string
	TaskType string
	Quorum   types.QuorumInfo
	Param    *StakeParam
	TxID     *ids.ID

	tx      *AddDelegatorTx
	signReq *types.SignRequest

	status       Status
	failed       bool
	StartTime    time.Time
	LastStepTime time.Time

	issuedByOthers bool
}

func NewAddDelegator(flowId string, quorum types.QuorumInfo, param *StakeParam) (*AddDelegator, error) {
	return &AddDelegator{
		FlowId:       flowId,
		TaskType:     taskTypeExport,
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
			t.logError(ctx, ErrMsgFailedToBuildAndSignTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToBuildAndSignTx)
		}
		t.status = StatusSignReqSent
	case StatusSignReqSent:
		return nil, errors.WithStack(t.getSignature(ctx))
	case StatusSignReqDone:
		return nil, errors.WithStack(t.sendTx(ctx))
	case StatusTxSent:
		status, err := t.checkTxStatus(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if status == core.TxStatusCommitted {
			t.status = StatusDone
			if t.issuedByOthers {
				t.logDebug(ctx, "tx committed, issued by others", []logger.Field{{"txId", t.TxID}}...)
			} else {
				prom.AddDelegatorTxCommitted.Inc()
				t.logDebug(ctx, "tx committed, issued by myself", []logger.Field{{"txId", t.TxID}}...)
			}
			return nil, nil
		}
		if t.issuedByOthers {
			t.logDebug(ctx, "tx issued by others, but uncommitted", []logger.Field{{"txId", t.TxID}, {"status", status.String()}}...)
		} else {
			t.logDebug(ctx, "tx issued by myself, but uncommitted", []logger.Field{{"txId", t.TxID}, {"status", status.String()}}...)
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
	prom.MpcSignPostedForAddDelegatorTx.Inc()
	t.logDebug(ctx, "sent signing request", []logger.Field{{"signReq", t.signReq.ReqID}}...)
	return nil
}

func (t *AddDelegator) getSignature(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().Result(context.Background(), t.signReq.ReqID)
	// TODO: Handle 404
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.Status != types.StatusDone {
		status := strings.ToLower(string(res.Status))
		if strings.Contains(status, "error") || strings.Contains(status, "err") {
			return t.failIfErrorf(err, "failed to sign AddDelegatorTx from P-Chain, status:%v", status)
		}
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.signReq.ReqID}, {"status", status}}...)
		return nil
	}

	prom.MpcSignDoneForAddDelegatorTx.Inc()
	sig := new(types.Signature).FromHex(res.Result)
	err = t.tx.SetTxSig(*sig)
	if err != nil {
		return t.failIfErrorf(err, "failed to set signature")
	}
	t.status = StatusSignReqDone

	_, err = t.tx.SignedTxBytes()
	if err != nil {
		return t.failIfErrorf(err, "failed to get signed bytes")
	}

	txID := t.tx.ID()
	t.TxID = &txID
	return nil
}

func (t *AddDelegator) sendTx(ctx core.TaskContext) error {
	// waits for arbitrary duration to elapse to reduce race condition.
	// TODO: tune the random number to a more suitable value
	rand.Seed(time.Now().UnixNano())
	random := rand.Int63n(5000)
	<-time.After(time.Millisecond * time.Duration(random))

	// check whether tx has been sent by other mpc-controller
	status, err := t.checkTxStatus(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// if tx status is unknown, it's reasonable to send one
	if status == core.TxStatusUnknown {
		signed, _ := t.tx.SignedTxBytes()
		_, err := ctx.IssuePChainTx(signed)
		if err != nil {
			// TODO: handle errors, including connection timeout
			status, err = t.checkTxStatus(ctx)
			if err != nil {
				return errors.WithStack(err)
			}
			t.logDebug(ctx, ErrMsgIssueTxFail, []logger.Field{{"txId", t.TxID}, {"status", status.String()}, {"error", err.Error()}}...)
			return errors.Wrapf(err, "%v, txId:%v", ErrMsgIssueTxFail, t.TxID)
		}

		prom.AddDelegatorTxIssued.Inc()
		t.status = StatusTxSent
		t.logDebug(ctx, "tx issued by myself", logger.Field{Key: "txId", Value: t.TxID})
		return nil
	}

	// otherwise, it has been handled by other mpc-controller, don't need to re-send it
	t.issuedByOthers = true
	t.status = StatusTxSent
	t.logDebug(ctx, "tx issued by others", []logger.Field{{"txId", t.TxID}, {"status", status.String()}}...)
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

	prom.MpcTxBuilt.With(prometheus.Labels{"flow": "initialStake", "chain": "pChain", "tx": "addDelegatorTx"}).Inc()

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

func (t *AddDelegator) checkTxStatus(ctx core.TaskContext) (core.TxStatus, error) {
	status, err := ctx.CheckPChainTx(*t.TxID)
	if err != nil {
		// TODO: review correctness of error handling
		var errTxStatusUndefined *taskcontext.ErrTypTxStatusInvalid
		if errors.As(err, &errTxStatusUndefined) {
			t.logError(ctx, ErrMsgTxFail, err, []logger.Field{{"txId", t.TxID}}...)
			return status, t.failIfErrorf(err, "%v, txId: %v", ErrMsgTxFail, t.TxID)
		}

		var errTxStatusAborted *taskcontext.ErrTypTxStatusAborted
		if errors.As(err, &errTxStatusAborted) {
			t.logError(ctx, ErrMsgTxFail, err, []logger.Field{{"txId", t.TxID}}...)
			return status, t.failIfErrorf(err, "%v, txId: %v", ErrMsgTxFail, t.TxID)
		}

		var errTxStatusDropped *taskcontext.ErrTypTxStatusDropped
		if errors.As(err, &errTxStatusDropped) {
			t.logError(ctx, ErrMsgTxFail, err, []logger.Field{{"txId", t.TxID}}...)
			return status, t.failIfErrorf(err, "%v, txId: %v", ErrMsgTxFail, t.TxID)
		}

		t.logDebug(ctx, ErrMsgCheckTxStatusFail, []logger.Field{{"txId", t.TxID}, {"error", err.Error()}}...)
		return status, errors.Wrapf(err, "%v, txId: %v", ErrMsgCheckTxStatusFail, t.TxID)
	}
	return status, nil
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

func (t *AddDelegator) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
