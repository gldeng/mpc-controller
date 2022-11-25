package c2p

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"sync/atomic"
	"time"
)

var (
	_ core.Task = (*ExportFromCChain)(nil)
)

type ExportFromCChain struct {
	Status      Status
	Id          string
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
	return fmt.Sprintf("ExportC(%v)", t.Id)
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

func NewExportFromCChain(id string, quorum types.QuorumInfo, amount big.Int) (*ExportFromCChain, error) {
	return &ExportFromCChain{
		Status:      StatusInit,
		Id:          id,
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
				atomic.AddInt32(&core.NonceConsumers, -1)
				return next, errors.Wrap(err, "failed to export from C-Chain")
			}
			if time.Now().Sub(t.StartTime) >= timeOut {
				atomic.AddInt32(&core.NonceConsumers, -1)
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
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToBuildAndSignTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToBuildAndSignTx)
		} else {
			t.Status = StatusSignReqSent
		}
	case StatusSignReqSent:
		err := t.getSignatureAndSendTx(ctx)
		if err != nil {
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToGetSignatureAndSendTx, err)
			return nil, t.failIfErrorf(err, ErrMsgFailedToGetSignatureAndSendTx)
		}
		if t.TxID != nil {
			ctx.GetLogger().Debugf("id %v ExportTx ID is %v", t.Id, t.TxID.String())
			t.Status = StatusTxSent
		}
	case StatusTxSent:
		status, err := ctx.CheckCChainTx(*t.TxID)
		ctx.GetLogger().Debugf("id %v ExportTx Status is %v", t.Id, status)
		if err != nil {
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToCheckStatus, err)
			return nil, err
		}
		if !core.IsPending(status) {
			t.Status = StatusDone
			prom.C2PExportTxCommitted.Inc()
			return nil, nil
		}
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
	req, err := t.buildSignReq(t.Id+"-export", txHash)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToCreateSignRequest)
	}
	t.SignRequest = req

	err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfErrorf(err, ErrMsgFailedToSendSignRequest)
	}
	ctx.GetLogger().Debugf("sent signing ExportTx request from C-Chain, requestID:%v", req.ReqID)
	return nil
}

func (t *ExportFromCChain) getSignatureAndSendTx(ctx core.TaskContext) error {
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
		ctx.GetLogger().Debugf("signing ExportTx from C-Chain not done, requestID:%v, status:%v", t.SignRequest.ReqID, status) // TODO: timeout and quit
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
	_, err = ctx.IssueCChainTx(signed.SignedBytes())
	if err != nil {
		// TODO: Better handling, if another participant already sent the tx, we can't send again. So err is not exception, but a normal outcome.
		//return t.failIfErrorf(err, ErrMsgFailedToIssueTx+fmt.Sprintf(" tx: %v", signed))
		return nil
	}
	prom.C2PExportTxIssued.Inc()
	return nil
}

func (t *ExportFromCChain) failIfErrorf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	return errors.Wrap(err, fmt.Sprintf(format, a...))
}
