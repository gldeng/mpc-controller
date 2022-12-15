package c2p

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
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
	SignRequest    *types.SignRequest
	Failed         bool
	StartTime      time.Time
	LastStepTime   time.Time
	issuedByOthers bool
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
			if t.issuedByOthers {
				t.logDebug(ctx, "tx committed, issued by others", []logger.Field{{"txId", t.TxID}}...)
			} else {
				prom.C2PImportTxCommitted.Inc()
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

func (t *ImportIntoPChain) SignedTx() (*txs.Tx, error) {
	return PackSignedImportTx(t.Tx, t.TxCred)
}

func (t *ImportIntoPChain) buildSignReq(id string, hash []byte) (*types.SignRequest, error) {
	partiPubKeys, genPubKey, err := t.Quorum.CompressKeys()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compress public keys")
	}
	return &types.SignRequest{
		ReqID:                  id,
		CompressedGenPubKeyHex: genPubKey,
		CompressedPartiPubKeys: partiPubKeys,
		Hash:                   bytes.BytesToHex(hash),
	}, nil
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

	err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToSendSignRequest)
	}
	prom.MpcSignPostedForC2PImportTx.Inc()
	t.logDebug(ctx, "sent signing request", logger.Field{"signReq", req.ReqID})
	return nil
}

func (t *ImportIntoPChain) getSignature(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().Result(context.Background(), t.SignRequest.ReqID)
	// TODO: Handle 404
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.Status != types.StatusDone {
		status := strings.ToLower(string(res.Status))
		if strings.Contains(status, "error") || strings.Contains(status, "err") {
			return t.failIfErrorf(err, "failed to sign ImportTx from P-Chain, status:%v", status)
		}
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.SignRequest.ReqID}, {"status", status}}...)
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
		signed, _ := t.SignedTx()
		_, err := ctx.IssuePChainTx(signed.Bytes())
		if err != nil {
			// TODO: handle errors, including connection timeout
			t.logDebug(ctx, ErrMsgIssueTxFail, []logger.Field{{"txId", t.TxID}, {"error", err.Error()}}...)
			return errors.Wrapf(err, "%v, txId:%v", ErrMsgIssueTxFail, t.TxID)
		}

		prom.C2PImportTxIssued.Inc()
		t.Status = StatusTxSent
		t.logDebug(ctx, "tx issued by myself", logger.Field{Key: "txId", Value: t.TxID})
		return nil
	}

	// otherwise, it has been handled by other mpc-controller, don't need to re-send it
	t.issuedByOthers = true
	t.Status = StatusTxSent
	t.logDebug(ctx, "tx issued by others", []logger.Field{{"txId", t.TxID}, {"status", status.String()}}...)
	return nil
}

func (t *ImportIntoPChain) checkTxStatus(ctx core.TaskContext) (core.TxStatus, error) {
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
