package p2c

import (
	"context"
	"fmt"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/ethereum/go-ethereum/common"
	"time"

	"github.com/avalido/mpc-controller/core/mpc"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/bytes"
	utilstime "github.com/avalido/mpc-controller/utils/time"
	"github.com/pkg/errors"
)

const (
	taskTypeImport = "importC"
)

var (
	_ core.Task = (*ImportIntoCChain)(nil)
)

type ImportIntoCChain struct {
	Status   Status
	FlowId   core.FlowId
	TaskType string
	Quorum   types.QuorumInfo

	SignedExportTx *txs.Tx
	To             common.Address
	Tx             *evm.UnsignedImportTx
	TxHash         []byte
	TxCred         *secp256k1fx.Credential
	TxID           *ids.ID
	SignRequest    *mpc.SignRequest
	Failed         bool
	StartTime      *time.Time
	LastStepTime   *time.Time
	issuedByOthers bool
}

func (t *ImportIntoCChain) GetId() string {
	return fmt.Sprintf("%v-importC", t.FlowId)
}

func (t *ImportIntoCChain) FailedPermanently() bool {
	return t.Failed
}

func (t *ImportIntoCChain) IsSequential() bool {
	return false
}

func (t *ImportIntoCChain) IsDone() bool {
	return t.Status == StatusDone
}

func NewImportIntoCChain(flowId core.FlowId, quorum types.QuorumInfo, signedExportTx *txs.Tx, to common.Address) (*ImportIntoCChain, error) {
	return &ImportIntoCChain{
		Status:         StatusInit,
		FlowId:         flowId,
		TaskType:       taskTypeImport,
		Quorum:         quorum,
		SignedExportTx: signedExportTx,
		To:             to,
		Tx:             nil,
		TxHash:         nil,
		TxCred:         nil,
		TxID:           nil,
		SignRequest:    nil,
		Failed:         false,
		StartTime:      nil,
		LastStepTime:   nil,
		issuedByOthers: false,
	}, nil
}

func (t *ImportIntoCChain) Next(ctx core.TaskContext) ([]core.Task, error) {
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
		prom.TaskTimeout.With(prometheus.Labels{"flow": "initialStake", "task": taskTypeImport}).Inc()
		return nil, errors.New(ErrMsgTimedOut)
	}
	defer func() {
		now := time.Now()
		t.LastStepTime = &now
	}()
	return t.run(ctx)
}

func (t *ImportIntoCChain) run(ctx core.TaskContext) ([]core.Task, error) {
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
			{"status", status.String()},
		}...)
	}
	return nil, nil
}

func (t *ImportIntoCChain) SignedTx() (*evm.Tx, error) {
	return PackSignedImportTx(t.Tx, t.TxCred)
}

func (t *ImportIntoCChain) buildSignReq(id string, hash []byte) (*mpc.SignRequest, error) {
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

func (t *ImportIntoCChain) buildAndSignTx(ctx core.TaskContext) error {
	builder := NewTxBuilder(ctx.GetNetwork())
	tx, err := builder.ImportIntoCChain(t.To, t.SignedExportTx, t.FlowId.RequestHash[:])
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

	prom.MpcTxBuilt.With(prometheus.Labels{"flow": "initialStake", "chain": "cChain", "tx": "importTx"}).Inc()

	_, err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToSendSignRequest)
	}
	prom.MpcSignPostedForP2CImportTx.Inc()
	t.logDebug(ctx, "sent signing request", logger.Field{"signReq", req.RequestId})
	return nil
}

func (t *ImportIntoCChain) getSignature(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().CheckResult(context.Background(), &mpc.CheckResultRequest{RequestId: t.SignRequest.RequestId})
	if res.RequestStatus == mpc.CheckResultResponse_ERROR {
		return t.failIfErrorf(err, "failed to sign ImportTx from P-Chain, status:%v", res.RequestStatus.String())
	}

	if res.RequestStatus != mpc.CheckResultResponse_DONE {
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.SignRequest.RequestId}, {"status", res.RequestStatus.String()}}...)
		return nil
	}

	prom.MpcSignDoneForP2CImportTx.Inc()
	t.logInfo(ctx, "signing done", []logger.Field{{"signReq", t.SignRequest.RequestId}}...)
	txCred, err := ValidateAndGetCred(t.TxHash, *new(types.Signature).FromHex(res.Result), t.Quorum.PChainAddress())
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToValidateCredential)
	}
	t.TxCred = txCred
	signed, err := t.SignedTx()
	if err != nil {
		return t.failIfErrorf(err, ErrMsgPrepareSignedTx)
	}
	txId := signed.ID()
	t.TxID = &txId

	t.Status = StatusSignedTxReady
	return nil
}

// sendTx sends a tx to avalanche network. Without a consensus mechanism among the participants, every partiticipant
// attempts to issue the tx. We do the following to mitigate the race condition:
//  1. delay a random duration before sending tx
//  2. check tx status on-chain in case other participants already send it, if already sent (i.e. the tx is known to
//     avalanche network already before sending tx
//  3. check tx status again after sending failed which may be caused by another participant sending the same tx
//     at the same time
func (t *ImportIntoCChain) sendTx(ctx core.TaskContext) error {
	// waits for arbitrary duration to elapse to reduce race condition.
	utilstime.RandomDelay(5000)

	isIssued, err := t.checkIfTxIssued(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if isIssued {
		return nil
	}

	signed, _ := t.SignedTx()
	_, err = ctx.IssueCChainTx(signed.SignedBytes())
	if err != nil {
		//_, err := t.checkIfTxIssued(ctx) // TODO: Fix this
		return errors.WithStack(err)
	}

	t.Status = StatusTxSent
	prom.P2CImportTxIssued.Inc()
	t.logDebug(ctx, "tx issued", []logger.Field{{Key: "txId", Value: t.TxID}}...)
	return nil
}

func (t *ImportIntoCChain) checkIfTxIssued(ctx core.TaskContext) (bool, error) {
	status, err := t.checkTxStatus(ctx)
	if err != nil {
		t.logError(ctx, ErrMsgCheckTxStatusFail, err, []logger.Field{{"txId", t.TxID}}...)
		return false, errors.Wrapf(err, "%v, txId: %v", ErrMsgCheckTxStatusFail, t.TxID)
	}

	t.logDebug(ctx, "checked tx status", []logger.Field{
		{"txId", t.TxID},
		{"status", status.String()},
	}...)

	defer func() {
		if status == core.TxStatusProcessing {
			t.Status = StatusTxSent
			prom.P2CImportTxIssued.Inc()
		}
	}()

	switch status {
	case core.TxStatusCommitted:
		return true, nil
	case core.TxStatusProcessing:
		return true, nil
	default:
		return false, nil
	}
}

func (t *ImportIntoCChain) checkTxStatus(ctx core.TaskContext) (core.TxStatus, error) {
	status, err := ctx.CheckCChainTx(*t.TxID)
	if err != nil {
		return status, errors.WithStack(err)
	}

	defer func() {
		if status == core.TxStatusCommitted {
			t.Status = StatusDone
			prom.P2CImportTxCommitted.Inc()
		}
	}()

	switch status {
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

func (t *ImportIntoCChain) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *ImportIntoCChain) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *ImportIntoCChain) logInfo(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Info(msg, allFields...)
}

func (t *ImportIntoCChain) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
