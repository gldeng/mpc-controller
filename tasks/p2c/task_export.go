package p2c

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/core/mpc"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
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
	taskTypeExport = "exportP"
)

var (
	_ core.Task = (*ExportFromPChain)(nil)
)

type ExportFromPChain struct {
	Status   Status
	FlowId   core.FlowId
	TaskType string

	Utxo avax.UTXO

	Quorum      types.QuorumInfo
	Tx          *txs.ExportTx
	TxHash      []byte
	TxCred      *secp256k1fx.Credential
	TxID        *ids.ID
	SignRequest *mpc.SignRequest
	Failed      bool
	StartTime   *time.Time
}

func (t *ExportFromPChain) GetId() string {
	return fmt.Sprintf("%v-exportP", t.FlowId)
}

func (t *ExportFromPChain) FailedPermanently() bool {
	return t.Failed
}

func (t *ExportFromPChain) IsSequential() bool {
	return true
}

func (t *ExportFromPChain) IsDone() bool {
	return t.Status == StatusDone
}

func NewExportFromPChain(flowId core.FlowId, quorum types.QuorumInfo, utxo avax.UTXO) (*ExportFromPChain, error) {
	return &ExportFromPChain{
		Status:      StatusInit,
		FlowId:      flowId,
		TaskType:    taskTypeExport,
		Utxo:        utxo,
		Quorum:      quorum,
		Tx:          nil,
		TxHash:      nil,
		TxCred:      nil,
		TxID:        nil,
		SignRequest: nil,
		Failed:      false,
	}, nil
}

func (t *ExportFromPChain) Next(ctx core.TaskContext) ([]core.Task, error) {
	if t.StartTime == nil {
		now := time.Now()
		t.StartTime = &now
	}

	timeout := 60 * time.Minute
	interval := 2 * time.Second
	timer := time.NewTimer(interval)
	defer timer.Stop()
	var next []core.Task
	var err error

	for {
		select {
		case <-timer.C:
			next, err = t.run(ctx)
			if t.IsDone() || t.Failed {
				return next, errors.Wrap(err, "failed to export from P-Chain")
			}
			if time.Now().Sub(*t.StartTime) >= timeout {
				prom.TaskTimeout.With(prometheus.Labels{"flow": "initialStake", "task": taskTypeExport}).Inc()
				return nil, errors.New(ErrMsgTimedOut)
			}

			timer.Reset(interval)
		}
	}
}

func (t *ExportFromPChain) SignedTx() (*txs.Tx, error) {
	return PackSignedExportTx(t.Tx, t.TxCred)
}

func (t *ExportFromPChain) run(ctx core.TaskContext) ([]core.Task, error) {
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
			{"status", status.String()}}...)
	}
	return nil, nil
}

func (t *ExportFromPChain) buildSignReq(id string, hash []byte) (*mpc.SignRequest, error) {
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

func (t *ExportFromPChain) buildAndSignTx(ctx core.TaskContext) error {
	builder := NewTxBuilder(ctx.GetNetwork())

	tx, err := builder.ExportFromPChain(t.Utxo)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToBuildTx)
	}
	t.Tx = tx
	txHash, err := ExportTxHash(tx)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToGetTxHash)
	}
	t.TxHash = txHash

	prom.MpcTxBuilt.With(prometheus.Labels{"flow": "initialStake", "chain": "pChain", "tx": "exportTx"}).Inc()

	req, err := t.buildSignReq(t.GetId(), txHash)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCreateSignRequest)
	}
	t.SignRequest = req

	t.logDebug(ctx, "Built sign request", []logger.Field{
		{"signReq", types.SignRequest{
			ReqID:                  req.RequestId,
			Hash:                   req.Hash,
			CompressedGenPubKeyHex: req.PublicKey,
			CompressedPartiPubKeys: req.ParticipantPublicKeys}}}...)

	_, err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToSendSignRequest)
	}
	prom.MpcSignPostedForC2PExportTx.Inc()
	t.logDebug(ctx, "sent signing request", logger.Field{"signReq", req.RequestId})
	return nil
}

func (t *ExportFromPChain) getSignature(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().CheckResult(context.Background(), &mpc.CheckResultRequest{RequestId: t.SignRequest.RequestId})
	// TODO: Handle 404
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}
	if res.RequestStatus == mpc.CheckResultResponse_ERROR {
		return t.failIfErrorf(err, "failed to sign ExportTx from C-Chain, status:%v", res.RequestStatus.String())
	}

	if res.RequestStatus != mpc.CheckResultResponse_DONE {
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.SignRequest.RequestId}, {"status", res.RequestStatus.String()}}...)
		return nil
	}
	prom.MpcSignDoneForC2PExportTx.Inc()
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
func (t *ExportFromPChain) sendTx(ctx core.TaskContext) error {
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
	_, err = ctx.IssuePChainTx(signed.Bytes())
	if err != nil {
		_, err := t.checkIfTxIssued(ctx)
		return errors.WithStack(err)
	}

	t.Status = StatusTxSent
	prom.P2CExportTxIssued.Inc()
	t.logDebug(ctx, "tx issued", []logger.Field{{Key: "txId", Value: t.TxID}}...)
	return nil
}

func (t *ExportFromPChain) checkIfTxIssued(ctx core.TaskContext) (bool, error) {
	status, err := t.checkTxStatus(ctx)
	if err != nil {
		t.logError(ctx, ErrMsgCheckTxStatusFail, err, []logger.Field{{"txId", t.TxID}}...)
		return false, errors.Wrapf(err, "%v, txId: %v", ErrMsgCheckTxStatusFail, t.TxID)
	}

	t.logDebug(ctx, "checked tx status", []logger.Field{
		{"txId", t.TxID},
		{"status", status.String()}}...)

	defer func() {
		if status == core.TxStatusProcessing {
			t.Status = StatusTxSent
			prom.P2CExportTxIssued.Inc()
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

func (t *ExportFromPChain) checkTxStatus(ctx core.TaskContext) (core.TxStatus, error) {
	status, err := ctx.CheckPChainTx(*t.TxID)
	if err != nil {
		return status.Code, errors.WithStack(err)
	}

	defer func() {
		if status.Code == core.TxStatusCommitted {
			t.Status = StatusDone
			prom.P2CExportTxCommitted.Inc()
		}
	}()

	switch status.Code {
	case core.TxStatusUnknown:
		return status.Code, nil
	case core.TxStatusCommitted:
		return status.Code, nil
	case core.TxStatusProcessing:
		return status.Code, nil
	default:
		return status.Code, t.failIfErrorf(errors.New(status.String()), ErrMsgTxFail)
	}
}

func (t *ExportFromPChain) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *ExportFromPChain) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *ExportFromPChain) logInfo(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Info(msg, allFields...)
}

func (t *ExportFromPChain) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
