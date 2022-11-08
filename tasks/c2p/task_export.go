package c2p

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
	"math/big"
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
	SignRequest *core.SignRequest
	Failed      bool
}

func (t *ExportFromCChain) GetId() string {
	return fmt.Sprintf("ExportC(%v)", t.Id)
}

func (t *ExportFromCChain) FailedPermanently() bool {
	return t.Failed
}

func (t *ExportFromCChain) RequiresNonce() bool {
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
		SignRequest: nil,
	}, nil
}

func (t *ExportFromCChain) Next(ctx core.TaskContext) ([]core.Task, error) {
	interval := 100 * time.Millisecond
	timer := time.NewTimer(interval)
	for {
		select {
		case <-timer.C:
			next, err := t.run(ctx)
			if err != nil || t.Status == StatusDone {
				return next, err
			} else {
				timer.Reset(interval)
			}
		}
	}

	return nil, nil
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
			return nil, err
		} else {
			t.Status = StatusSignReqSent
		}
	case StatusSignReqSent:
		err := t.getSignatureAndSendTx(ctx)
		if err != nil {
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToGetSignatureAndSendTx, err)
			return nil, err
		} else {
			if t.TxID != nil {
				ctx.GetLogger().Debug(fmt.Sprintf("id %v ExportTx ID is %v", t.Id, t.TxID.String()))
			}

			t.Status = StatusTxSent
		}
	case StatusTxSent:
		status, err := ctx.CheckCChainTx(*t.TxID)
		ctx.GetLogger().Debug(fmt.Sprintf("id %v ExportTx Status is %v", t.Id, status))
		if err != nil {
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToCheckStatus, err)
			return nil, err
		}
		if !core.IsPending(status) {
			t.Status = StatusDone
			return nil, nil
		}
	}
	return nil, nil
}

func (t *ExportFromCChain) buildSignReq(id string, hash []byte) (*core.SignRequest, error) {
	var participantPks []string
	for _, pk := range t.Quorum.ParticipantPubKeys {
		participantPks = append(participantPks, hex.EncodeToString(pk))
	}
	return &core.SignRequest{
		ReqID:                  id,
		CompressedGenPubKeyHex: hex.EncodeToString(t.Quorum.PubKey),
		CompressedPartiPubKeys: participantPks,
		Hash:                   bytes.BytesToHex(hash),
	}, nil
}

func (t *ExportFromCChain) buildAndSignTx(ctx core.TaskContext) error {
	builder := NewTxBuilder(ctx.GetNetwork())
	nonce, err := ctx.NonceAt(t.Quorum.CChainAddress())
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToGetNonce)
	}
	amount, err := ToGwei(&t.Amount)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToConvertAmount)
	}
	tx, err := builder.ExportFromCChain(t.Quorum.PubKey, amount, nonce)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToBuildTx)
	}
	t.Tx = tx
	txHash, err := ExportTxHash(tx)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToGetTxHash)
	}
	t.TxHash = txHash
	req, err := t.buildSignReq(t.Id+"/export", txHash)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToCreateSignRequest)
	}
	t.SignRequest = req

	err = ctx.GetMpcClient().Sign(context.Background(), req)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToSendSignRequest)
	}
	return nil
}

func (t *ExportFromCChain) getSignatureAndSendTx(ctx core.TaskContext) error {
	res, err := ctx.GetMpcClient().Result(context.Background(), t.SignRequest.ReqID)
	// TODO: Handle 404
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToCheckSignRequest)
	}

	if res.Status != core.StatusDone {
		ctx.GetLogger().Debug(DebugMsgSignRequestNotDone)
		return nil
	}
	txCred, err := ValidateAndGetCred(t.TxHash, *new(events.Signature).FromHex(res.Result), t.Quorum.PChainAddress())
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToValidateCredential)
	}
	t.TxCred = txCred
	signed, err := t.SignedTx()
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToPrepareSignedTx)
	}
	_, err = ctx.IssueCChainTx(signed.SignedBytes())
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToIssueTx)
	}
	txId := signed.ID()
	t.TxID = &txId
	return nil
}

func (t *ExportFromCChain) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	msg = fmt.Sprintf("[%v] %v", t.Id, msg)
	return errors.Wrap(err, msg)
}
