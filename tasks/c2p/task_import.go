package c2p

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
)

var (
	_ core.Task = (*ImportIntoPChain)(nil)
)

type ImportIntoPChain struct {
	Status Status
	Id     string
	Quorum types.QuorumInfo

	SignedExportTx *evm.Tx
	Tx             *txs.ImportTx
	TxHash         []byte
	TxCred         *secp256k1fx.Credential
	TxID           *ids.ID
	SignRequest    *core.SignRequest
	Failed         bool
}

func (t *ImportIntoPChain) GetId() string {
	return fmt.Sprintf("ImportP(%v)", t.Id)
}

func (t *ImportIntoPChain) FailedPermanently() bool {
	return t.Failed
}

func (t *ImportIntoPChain) RequiresNonce() bool {
	return false
}

func (t *ImportIntoPChain) IsDone() bool {
	return t.Status == StatusDone
}

func NewImportIntoPChain(id string, quorum types.QuorumInfo, signedExportTx *evm.Tx) (*ImportIntoPChain, error) {
	return &ImportIntoPChain{
		Status:         StatusInit,
		Id:             id,
		Quorum:         quorum,
		SignedExportTx: signedExportTx,
		Tx:             nil,
		TxHash:         nil,
		TxCred:         nil,
		TxID:           nil,
		SignRequest:    nil,
	}, nil
}

func (t *ImportIntoPChain) Next(ctx core.TaskContext) ([]core.Task, error) {
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
				ctx.GetLogger().Debug(fmt.Sprintf("id %v ImportTx ID is %v", t.Id, t.TxID.String()))
			}

			t.Status = StatusTxSent
		}
	case StatusTxSent:
		status, err := ctx.CheckPChainTx(*t.TxID)
		ctx.GetLogger().Debug(fmt.Sprintf("id %v ImportTx Status is %v", t.Id, status))
		if err != nil {
			ctx.GetLogger().Errorf("%v, error:%+v", ErrMsgFailedToCheckStatus, err)
		}
		if !core.IsPending(status) {
			t.Status = StatusDone
			return nil, nil
		}
	}
	return nil, nil
}

func (t *ImportIntoPChain) SignedTx() (*txs.Tx, error) {
	return PackSignedImportTx(t.Tx, t.TxCred)
}

func (t *ImportIntoPChain) buildSignReq(id string, hash []byte) (*core.SignRequest, error) {
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

func (t *ImportIntoPChain) buildAndSignTx(ctx core.TaskContext) error {
	builder := NewTxBuilder(ctx.GetNetwork())
	tx, err := builder.ImportIntoPChain(t.Quorum.PubKey, t.SignedExportTx)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToBuildAndSignTx)
	}
	t.Tx = tx
	txHash, err := ImportTxHash(tx)
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToGetTxHash)
	}
	t.TxHash = txHash
	req, err := t.buildSignReq(t.Id+"/import", txHash)
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

func (t *ImportIntoPChain) getSignatureAndSendTx(ctx core.TaskContext) error {
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
	txId := signed.ID()
	_, err = ctx.IssuePChainTx(signed.Bytes()) // If it's dropped, no ID will be returned?
	if err != nil {
		return t.failIfError(err, ErrMsgFailedToIssueTx)
	}
	t.TxID = &txId
	return nil
}

func (t *ImportIntoPChain) failIfError(err error, msg string) error {
	if err == nil {
		return nil
	}
	t.Failed = true
	msg = fmt.Sprintf("[%v] %v", t.Id, msg)
	return errors.Wrap(err, msg)
}
