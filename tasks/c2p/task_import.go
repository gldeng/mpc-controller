package c2p

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/core/mpc"
	"time"

	"github.com/avalido/mpc-controller/taskcontext"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
)

const (
	taskTypeImport = "import"
)

var (
	_ core.Task = (*ImportIntoPChain)(nil)
)

type ImportIntoPChain struct {
	Status   Status
	FlowId   string
	TaskType string
	Quorum   types.QuorumInfo

	SignedExportTx *evm.Tx
	Tx             *txs.ImportTx
	TxHash         []byte
	TxCred         *secp256k1fx.Credential
	TxID           *ids.ID
	SignRequest    *mpc.SignRequest
	Failed         bool
	StartTime      time.Time
	LastStepTime   time.Time
}

func (t *ImportIntoPChain) GetId() string {
	return fmt.Sprintf("%v-import", t.FlowId)
}

func (t *ImportIntoPChain) FailedPermanently() bool {
	return t.Failed
}

func (t *ImportIntoPChain) IsSequential() bool {
	return false
}

func (t *ImportIntoPChain) IsDone() bool {
	return t.Status == StatusDone
}

func NewImportIntoPChain(flowId string, quorum types.QuorumInfo, signedExportTx *evm.Tx) (*ImportIntoPChain, error) {
	return &ImportIntoPChain{
		Status:         StatusInit,
		FlowId:         flowId,
		TaskType:       taskTypeImport,
		Quorum:         quorum,
		SignedExportTx: signedExportTx,
		Tx:             nil,
		TxHash:         nil,
		TxCred:         nil,
		TxID:           nil,
		SignRequest:    nil,
		Failed:         false,
		StartTime:      time.Now(),
		LastStepTime:   time.Now(),
	}, nil
}

func (t *ImportIntoPChain) Next(ctx core.TaskContext) ([]core.Task, error) {
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

func (t *ImportIntoPChain) run(ctx core.TaskContext) ([]core.Task, error) {
	switch t.Status {
	case StatusInit:
		err := t.buildAndSignTx(ctx)
		if err != nil {
			t.logError(ctx, ErrMsgFailedToBuildAndSignTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToBuildAndSignTx)
		} else {
			t.Status = StatusSignReqSent
		}
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
			t.Status = StatusDone
			prom.C2PImportTxCommitted.Inc()
			t.logDebug(ctx, "tx committed", []logger.Field{{"txId", t.TxID}}...)
			return nil, nil
		}
		t.logDebug(ctx, "tx issued but uncommitted", []logger.Field{{"txId", t.TxID}, {"status", status}}...)
	}
	return nil, nil
}

func (t *ImportIntoPChain) SignedTx() (*txs.Tx, error) {
	return PackSignedImportTx(t.Tx, t.TxCred)
}

func (t *ImportIntoPChain) buildSignReq(id string, hash []byte) (*mpc.SignRequest, error) {
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

func (t *ImportIntoPChain) buildAndSignTx(ctx core.TaskContext) error {
	builder := NewTxBuilder(ctx.GetNetwork())
	tx, err := builder.ImportIntoPChain(t.Quorum.PubKey, t.SignedExportTx)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToBuildAndSignTx)
	}
	t.Tx = tx
	txHash, err := ImportTxHash(tx)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToGetTxHash)
	}
	t.TxHash = txHash
	req, err := t.buildSignReq(t.GetId(), txHash)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCreateSignRequest)
	}

	t.SignRequest = req

	prom.MpcTxBuilt.With(prometheus.Labels{"flow": "initialStake", "chain": "pChain", "tx": "importTx"}).Inc()

	_, err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToSendSignRequest)
	}
	prom.MpcSignPostedForC2PImportTx.Inc()
	t.logDebug(ctx, "sent signing request", logger.Field{"signReq", req.RequestId})
	return nil
}

func (t *ImportIntoPChain) getSignatureAndSendTx(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().CheckResult(context.Background(), &mpc.CheckResultRequest{RequestId: t.SignRequest.RequestId})
	// TODO: Handle 404
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.RequestStatus == mpc.CheckResultResponse_ERROR {
		return t.failIfErrorf(err, "failed to sign ImportTx from P-Chain, status:%v", res.RequestStatus.String())
	}

	if res.RequestStatus != mpc.CheckResultResponse_DONE {
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.SignRequest.RequestId}, {"status", res.RequestStatus.String()}}...)
		return nil
	}
	prom.MpcSignDoneForC2PImportTx.Inc()
	txCred, err := ValidateAndGetCred(t.TxHash, *new(types.Signature).FromHex(res.Result), t.Quorum.PChainAddress())
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToValidateCredential)
	}
	t.TxCred = txCred
	signed, err := t.SignedTx()
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToPrepareSignedTx)
	}
	txId := signed.ID()
	t.TxID = &txId
	// TODO: check tx status before issuing, which may has been committed by other mpc-controller?
	_, err = ctx.IssuePChainTx(signed.Bytes()) // If it's dropped, no ID will be returned?
	if err != nil {
		// TODO: Better handling, if another participant already sent the tx, we can't send again. So err is not exception, but a normal outcome.
		//var errInputsInvalid *taskcontext.ErrTypeTxInputsInvalid
		//var errUTXOConsumedFail *taskcontext.ErrTypeUTXOConsumeFail
		//if errors.As(err, &errInputsInvalid) || errors.As(err, errUTXOConsumedFail) {
		//	t.logError(ctx, ErrMsgTxFail, err, logger.Field{Key: "txId", Value: t.TxID})
		//	return t.failIfErrorf(err, "%v, txId:%v", ErrMsgTxFail, t.TxID)
		//}
		//t.logDebug(ctx, ErrMsgIssueTxFail, []logger.Field{{"txId", t.TxID}, {"error", err.Error()}}...)
		//return errors.Wrapf(err, "%v, txId:%v", ErrMsgIssueTxFail, t.TxID)
		t.logDebug(ctx, ErrMsgIssueTxFail, []logger.Field{{"txId", t.TxID}, {"error", err.Error()}}...)
		return nil
	}
	prom.C2PImportTxIssued.Inc()
	t.logDebug(ctx, "issued tx", logger.Field{Key: "txID", Value: t.TxID})
	return nil
}

func (t *ImportIntoPChain) getSignature(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().CheckResult(context.Background(), &mpc.CheckResultRequest{RequestId: t.SignRequest.RequestId})
	if res.RequestStatus == mpc.CheckResultResponse_ERROR {
		return t.failIfErrorf(err, "failed to sign ImportTx from P-Chain, status:%v", res.RequestStatus.String())
	}

	if res.RequestStatus != mpc.CheckResultResponse_DONE {
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.SignRequest.RequestId}, {"status", res.RequestStatus.String()}}...)
		return nil
	}

	prom.MpcSignDoneForC2PImportTx.Inc()
	txCred, err := ValidateAndGetCred(t.TxHash, *new(types.Signature).FromHex(res.Result), t.Quorum.PChainAddress())
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToValidateCredential)
	}
	t.TxCred = txCred
	t.Status = StatusSignReqDone

	signed, err := t.SignedTx()
	if err != nil {
		return t.failIfErrorf(err, ErrMsgPrepareSignedTx)
	}
	txId := signed.ID()
	t.TxID = &txId
	return nil
}

func (t *ImportIntoPChain) sendTx(ctx core.TaskContext) error {
	signed, _ := t.SignedTx()
	_, err := ctx.IssuePChainTx(signed.Bytes())
	if err != nil {
		// TODO: handle errors, including connection timeout
		t.logDebug(ctx, ErrMsgIssueTxFail, []logger.Field{{"txId", t.TxID}, {"error", err.Error()}}...)
	}

	prom.C2PImportTxIssued.Inc()
	t.Status = StatusTxSent
	t.logDebug(ctx, "issued tx", logger.Field{Key: "txId", Value: t.TxID})
	return nil
}

func (t *ImportIntoPChain) checkTxStatus(ctx core.TaskContext) (core.TxStatus, error) {
	status, err := ctx.CheckPChainTx(*t.TxID)
	if err != nil {
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

func (t *ImportIntoPChain) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *ImportIntoPChain) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *ImportIntoPChain) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
