package c2p

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/ids"
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
	taskTypeExport = "export"
)

var (
	_ core.Task = (*ExportFromCChain)(nil)
)

type ExportFromCChain struct {
	Status      Status
	FlowId      string
	TaskType    string
	Amount      big.Int
	Quorum      types.QuorumInfo
	Tx          *evm.UnsignedExportTx
	TxHash      []byte
	TxCred      *secp256k1fx.Credential
	TxID        *ids.ID
	SignRequest *types.SignRequest
	Failed      bool
	StartTime   time.Time
}

func (t *ExportFromCChain) GetId() string {
	return fmt.Sprintf("%v-export", t.FlowId)
}

func (t *ExportFromCChain) FailedPermanently() bool {
	return t.Failed
}

func (t *ExportFromCChain) IsSequential() bool {
	return true
}

func (t *ExportFromCChain) IsDone() bool {
	return t.Status == StatusDone
}

func NewExportFromCChain(flowId string, quorum types.QuorumInfo, amount big.Int) (*ExportFromCChain, error) {
	return &ExportFromCChain{
		Status:      StatusInit,
		FlowId:      flowId,
		TaskType:    taskTypeExport,
		Amount:      amount,
		Quorum:      quorum,
		Tx:          nil,
		TxHash:      nil,
		TxCred:      nil,
		TxID:        nil,
		SignRequest: nil,
		Failed:      false,
		StartTime:   time.Now(),
	}, nil
}

func (t *ExportFromCChain) Next(ctx core.TaskContext) ([]core.Task, error) {
	timeOut := 30 * time.Minute
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
				return next, errors.Wrap(err, "failed to export from C-Chain")
			}
			if time.Now().Sub(t.StartTime) >= timeOut {
				return nil, errors.New(ErrMsgTimedOut)
			}

			timer.Reset(interval)
		}
	}
}

func (t *ExportFromCChain) SignedTx() (*evm.Tx, error) {
	return PackSignedExportTx(t.Tx, t.TxCred)
}

func (t *ExportFromCChain) run(ctx core.TaskContext) ([]core.Task, error) {
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

func (t *ExportFromCChain) buildSignReq(id string, hash []byte) (*types.SignRequest, error) {
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

func (t *ExportFromCChain) buildAndSignTx(ctx core.TaskContext) error {
	builder := NewTxBuilder(ctx.GetNetwork())
	nonce, err := ctx.NonceAt(t.Quorum.CChainAddress())
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToGetNonce)
	}
	t.logDebug(ctx, "got nonce", []logger.Field{{"nonce", nonce}, {"address", t.Quorum.CChainAddress()}}...)
	amount, err := ToGwei(&t.Amount)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToConvertAmount)
	}
	tx, err := builder.ExportFromCChain(t.Quorum.PubKey, amount, nonce)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToBuildTx)
	}
	t.Tx = tx
	txHash, err := ExportTxHash(tx)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToGetTxHash)
	}
	t.TxHash = txHash

	prom.MpcTxBuilt.With(prometheus.Labels{"flow": "initialStake", "chain": "cChain", "tx": "exportTx"}).Inc()

	req, err := t.buildSignReq(t.GetId(), txHash)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCreateSignRequest)
	}
	t.SignRequest = req

	err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToSendSignRequest)
	}
	prom.MpcSignPostedForC2PExportTx.Inc()
	t.logDebug(ctx, "sent signing request", logger.Field{"signReq", req.ReqID})
	return nil
}

func (t *ExportFromCChain) getSignature(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().Result(context.Background(), t.SignRequest.ReqID)
	// TODO: Handle 404
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.Status != types.StatusDone {
		status := strings.ToLower(string(res.Status))
		if strings.Contains(status, "error") || strings.Contains(status, "err") {
			return t.failIfErrorf(err, "failed to sign ExportTx from C-Chain, status:%v", status)
		}
		t.logDebug(ctx, "signing not done", []logger.Field{{"signReq", t.SignRequest.ReqID}, {"status", status}}...)
		return nil
	}
	prom.MpcSignDoneForC2PExportTx.Inc()
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

func (t *ExportFromCChain) sendTx(ctx core.TaskContext) error {
	// waits for arbitrary duration to elapse to reduce race condition.
	// TODO: tune the random number to a more suitable value
	rand.Seed(time.Now().UnixNano())
	random := rand.Int63n(5000)
	<-time.After(time.Millisecond * time.Duration(random))

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
		_, err := t.checkIfTxIssued(ctx)
		return errors.WithStack(err)
	}

	t.Status = StatusTxSent
	prom.C2PExportTxIssued.Inc()
	t.logDebug(ctx, "tx issued", []logger.Field{{Key: "txId", Value: t.TxID}}...)
	return nil
}

func (t *ExportFromCChain) checkIfTxIssued(ctx core.TaskContext) (bool, error) {
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
			prom.C2PExportTxIssued.Inc()
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

func (t *ExportFromCChain) checkTxStatus(ctx core.TaskContext) (core.TxStatus, error) {
	status, err := ctx.CheckCChainTx(*t.TxID)
	if err != nil {
		return status, errors.WithStack(err)
	}

	defer func() {
		if status == core.TxStatusCommitted {
			t.Status = StatusDone
			prom.C2PExportTxCommitted.Inc()
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

func (t *ExportFromCChain) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}

func (t *ExportFromCChain) logDebug(ctx core.TaskContext, msg string, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+2)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	ctx.GetLogger().Debug(msg, allFields...)
}

func (t *ExportFromCChain) logError(ctx core.TaskContext, msg string, err error, fields ...logger.Field) {
	allFields := make([]logger.Field, 0, len(fields)+3)
	allFields = append(allFields, logger.Field{"flowId", t.FlowId})
	allFields = append(allFields, logger.Field{"taskType", t.TaskType})
	allFields = append(allFields, fields...)
	allFields = append(allFields, logger.Field{"error", fmt.Sprintf("%+v", err)})
	ctx.GetLogger().Error(msg, allFields...)
}
