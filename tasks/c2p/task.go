package c2p

import (
	"context"
	"encoding/hex"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/pkg/errors"
	"math/big"
)

var (
	_ pool.Task = (*TransferC2P)(nil)
)

const (
	StatusStarted Status = iota
	StatusBuilt

	StatusExportTxSigningPosted
	StatusExportTxSigningDone
	StatusExportTxIssued
	StatusExportTxApproved
	StatusExportTxFailed

	StatusImportTxSigningPosted
	StatusImportTxSigningDone
	StatusImportTxIssued
	StatusImportTxApproved
	StatusImportTxFailed

	StatusNewSignReqSent
	StatusNewTxSent
	StatusNewDone
)

type Status int

type TransferC2P struct {
	Status          Status
	Id              string
	Amount          big.Int
	Quorum          QuorumInfo
	Txs             *Txs
	SignReqs        []*core.SignRequest
	ExportIssueTx   *txissuer.Tx
	ExportTxStatus  pool.Status
	ExportTxSignRes *core.Result

	ImportIssueTx   *txissuer.Tx
	ImportTxStatus  pool.Status
	ImportTxSignRes *core.Result
}

func (t *TransferC2P) IsDone() bool {
	return t.Status == StatusImportTxApproved || t.Status == StatusImportTxFailed
}

func New(id string, amount big.Int, quorum QuorumInfo) (*TransferC2P, error) {
	return &TransferC2P{
		Status:          0,
		Id:              id,
		Amount:          amount,
		Quorum:          quorum,
		Txs:             nil,
		SignReqs:        nil,
		ExportIssueTx:   nil,
		ExportTxSignRes: nil,
		ImportIssueTx:   nil,
		ImportTxSignRes: nil,
	}, nil
}

func (t *TransferC2P) Next(ctx pool.TaskContext) ([]pool.Task, error) {
	self := []pool.Task{t}
	switch t.Status {
	case StatusStarted:
		err := t.buildTask(ctx)
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, "Failed to build task")
			return nil, err
		}
		t.Status = StatusBuilt
	case StatusBuilt:
		err := ctx.GetMpcClient().Sign(context.Background(), t.SignReqs[0])
		ctx.GetLogger().ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusExportTxSigningPosted
		}
	case StatusExportTxSigningPosted:
		res, err := ctx.GetMpcClient().Result(context.Background(), t.SignReqs[0].ReqID)
		ctx.GetLogger().ErrorOnError(err, "Failed to check signing result")

		if res.Status != core.StatusDone {
			ctx.GetLogger().Debug("Signing task not done")
			return self, nil
		}
		t.Status = StatusExportTxSigningDone
		t.ExportTxSignRes = res
	case StatusExportTxSigningDone:
		sig := new(events.Signature).FromHex(t.ExportTxSignRes.Result)
		err := t.Txs.SetExportTxSig(*sig)
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, "Failed to set signature")
			return nil, nil
		}

		signedBytes, err := t.Txs.SignedExportTxBytes()
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, "Failed to get signed bytes")
			return nil, nil
		}

		tx := txissuer.Tx{
			ReqID: t.SignReqs[0].ReqID,
			TxID:  t.Txs.ExportTxID(),
			Chain: txissuer.ChainC,
			Bytes: signedBytes,
		}
		t.ExportIssueTx = &tx

		_, err = ctx.IssueCChainTx(t.ExportIssueTx.Bytes)
		if err == nil {
			t.Status = StatusExportTxIssued
		}
	case StatusExportTxIssued:
		status, err := ctx.CheckCChainTx(t.ExportIssueTx.TxID)
		t.ExportTxStatus = status
		if err == nil && pool.IsFailed(status) {
			t.Status = StatusExportTxFailed
		}

		if err == nil && pool.IsSuccessful(status) {
			t.Status = StatusExportTxApproved
		}
	case StatusExportTxFailed:
		fallthrough
	case StatusExportTxApproved:
		signReq, err := t.buildImportTxSignReq(t.Txs)
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, "failed to build ImportTx signing request")
			return nil, nil
		}

		t.SignReqs[1] = signReq

		err = ctx.GetMpcClient().Sign(context.Background(), t.SignReqs[1])
		ctx.GetLogger().ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusImportTxSigningPosted
		}
	case StatusImportTxSigningPosted:
		res, err := ctx.GetMpcClient().Result(context.Background(), t.SignReqs[1].ReqID)
		ctx.GetLogger().ErrorOnError(err, "Failed to check signing result")

		if res.Status != core.StatusDone {
			ctx.GetLogger().Debug("Signing task not done")
			return self, nil
		}
		t.Status = StatusImportTxSigningDone
		t.ImportTxSignRes = res
	case StatusImportTxSigningDone:
		sig := new(events.Signature).FromHex(t.ImportTxSignRes.Result)
		err := t.Txs.SetImportTxSig(*sig)
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, "Failed to set signature")
			return nil, nil
		}

		signedBytes, err := t.Txs.SignedImportTxBytes()
		if err != nil {
			ctx.GetLogger().ErrorOnError(err, "Failed to get signed bytes")
			return nil, nil
		}

		tx := txissuer.Tx{
			ReqID: t.SignReqs[1].ReqID,
			TxID:  t.Txs.ImportTxID(),
			Chain: txissuer.ChainP,
			Bytes: signedBytes,
		}
		t.ImportIssueTx = &tx

		_, err = ctx.IssuePChainTx(t.ImportIssueTx.Bytes)
		if err == nil {
			t.Status = StatusImportTxIssued
		}
	case StatusImportTxIssued:
		status, err := ctx.CheckPChainTx(t.ImportIssueTx.TxID)
		t.ImportTxStatus = status
		if err == nil && pool.IsFailed(t.ImportTxStatus) {
			t.Status = StatusImportTxFailed
		}

		if err == nil && t.ImportIssueTx.Status == txissuer.StatusApproved {
			t.Status = StatusImportTxApproved
		}
	case StatusImportTxFailed:
		fallthrough
	case StatusImportTxApproved:
		evt := ImportedEvent{Tx: t.Txs.importTx}
		ctx.Emit(&evt)
		ctx.GetLogger().Info("Stake atomic task handled", []logger.Field{{"stakeAtomicTaskHandled", evt}}...)
		return nil, nil
	}
	return self, nil
}

func (t *TransferC2P) buildTask(ctx pool.TaskContext) error {
	txs, err := t.buildTxs(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to build txs")
	}

	signReq, err := t.buildExportTxSignReq(txs)
	if err != nil {
		return errors.Wrapf(err, "failed to build ExportTx signing request")
	}

	t.Txs = txs
	t.SignReqs = make([]*core.SignRequest, 2)
	t.SignReqs[0] = signReq
	return nil
}

func (t *TransferC2P) buildExportTxSignReq(txs *Txs) (*core.SignRequest, error) {
	exportTxHash, err := txs.ExportTxHash()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ExportTx hash")
	}

	return t.buildSignReq(t.Id+"/export", exportTxHash)
}

func (t *TransferC2P) buildImportTxSignReq(txs *Txs) (*core.SignRequest, error) {
	importTxHash, err := txs.ImportTxHash()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ImportTx hash")
	}

	return t.buildSignReq(t.Id+"/import", importTxHash)
}

func (t *TransferC2P) buildSignReq(id string, hash []byte) (*core.SignRequest, error) {
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

func (t *TransferC2P) buildTxs(ctx pool.TaskContext) (*Txs, error) {

	nAVAXAmount := new(big.Int).Div(&t.Amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() {
		return nil, errors.New(ErrMsgInvalidUint64)
	}

	cChainAddr, _ := (storage.PubKey(t.Quorum.PubKey)).CChainAddress()
	pChainAddr, _ := (storage.PubKey(t.Quorum.PubKey)).PChainAddress()
	nonce, _ := ctx.NonceAt(cChainAddr)

	st := Txs{
		Nonce:         nonce,
		DelegateAmt:   nAVAXAmount.Uint64(),
		CChainAddress: cChainAddr,
		PChainAddress: pChainAddr,

		BaseFeeGwei: cchain.BaseFeeGwei,
		NetworkID:   ctx.GetNetwork().NetworkID(),
		CChainID:    ctx.GetNetwork().CChainID(),
		Asset:       ctx.GetNetwork().Asset(),
		ImportFee:   ctx.GetNetwork().ImportFee(),
	}

	return &st, nil
}
