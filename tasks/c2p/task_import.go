package c2p

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		err := t.getSignatureAndSendTx(ctx)
		if err != nil {
			t.logError(ctx, ErrMsgFailedToGetSignatureAndSendTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToGetSignatureAndSendTx)
		}
		if t.TxID != nil {
			t.logDebug(ctx, "sent importTx", logger.Field{"importTx", t.TxID})
			t.Status = StatusTxSent
		}
	case StatusTxSent:
		status, err := ctx.CheckPChainTx(*t.TxID)
		t.logDebug(ctx, "checked importTx status", []logger.Field{{"importTx", t.TxID}, {"status", status}}...)
		if err != nil {
			t.logError(ctx, ErrMsgFailedToCheckStatus, err, logger.Field{"importTx", t.TxID})
		}
		if !core.IsPending(status) {
			t.Status = StatusDone
			prom.C2PImportTxCommitted.Inc()
			return nil, nil
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

	err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToSendSignRequest)
	}
	t.logDebug(ctx, "sent signing request", logger.Field{"signReq", req.ReqID})
	return nil
}

func (t *ImportIntoPChain) getSignatureAndSendTx(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().Result(context.Background(), t.SignRequest.ReqID)
	// TODO: Handle 404
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.Status != types.StatusDone {
		status := strings.ToLower(string(res.Status))
		if strings.Contains(status, "error") || strings.Contains(status, "err") {
			return t.failIfErrorf(err, "failed to sign ImportTx into P-Chain, status:%v", status)
		}
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.SignRequest.ReqID}, {"status", status}}...)
		return nil
	}
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
		//return t.failIfErrorf(err, ErrMsgFailedToIssueTx)
		return nil
	}
	prom.C2PImportTxIssued.Inc()
	return nil
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
