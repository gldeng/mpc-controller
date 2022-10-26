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
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/utils/bytes"
)

var (
	_ pool.Task = (*ImportIntoPChain)(nil)
)

type ImportIntoPChain struct {
	Status Status
	Id     string
	Quorum QuorumInfo

	SignedExportTx *evm.Tx
	Tx             *txs.ImportTx
	TxHash         []byte
	TxCred         *secp256k1fx.Credential
	TxID           *ids.ID
	SignRequest    *core.SignRequest
}

func (t *ImportIntoPChain) IsDone() bool {
	return t.Status == StatusNewDone
}

func NewImportIntoPChain(id string, quorum QuorumInfo, signedExportTx *evm.Tx) (*ImportIntoPChain, error) {
	return &ImportIntoPChain{
		Status:         StatusStarted,
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

func (t *ImportIntoPChain) Next(ctx pool.TaskContext) ([]pool.Task, error) {
	self := []pool.Task{t}
	switch t.Status {
	case StatusStarted:
		builder := NewTxBuilder(ctx.GetNetwork())
		tx, err := builder.ImportIntoPChain(t.Quorum.PubKey, t.SignedExportTx)
		ctx.GetLogger().ErrorOnError(err, "failed to build import tx")
		t.Tx = tx
		txHash, err := ImportTxHash(tx)
		t.TxHash = txHash
		ctx.GetLogger().ErrorOnError(err, "failed to get import tx hash")
		req, err := t.buildSignReq(t.Id+"/import", txHash)
		t.SignRequest = req
		ctx.GetLogger().ErrorOnError(err, "failed create sign request")
		err = ctx.GetMpcClient().Sign(context.Background(), req)
		ctx.GetLogger().ErrorOnError(err, "failed to post signing request")
		if err == nil {
			t.Status = StatusNewSignReqSent
		}
	case StatusNewSignReqSent:
		res, err := ctx.GetMpcClient().Result(context.Background(), t.SignRequest.ReqID)
		// TODO: Handle 404
		ctx.GetLogger().ErrorOnError(err, "Failed to check signing result")

		if res.Status != core.StatusDone {
			ctx.GetLogger().Debug("Signing task not done")
			return self, nil
		}
		txCred, err := ValidateAndGetCred(t.TxHash, *new(events.Signature).FromHex(res.Result), t.Quorum.PChainAddress())
		ctx.GetLogger().ErrorOnError(err, "failed to validate cred")
		t.TxCred = txCred
		signed, err := t.SignedTx()
		ctx.GetLogger().ErrorOnError(err, "failed to get signed tx")
		txId := signed.ID()
		_, err = ctx.IssuePChainTx(signed.Bytes()) // If it's dropped, no ID will be returned?
		ctx.GetLogger().ErrorOnError(err, "failed to issue tx")
		t.TxID = &txId
		t.Status = StatusNewTxSent
	case StatusNewTxSent:
		status, err := ctx.CheckPChainTx(*t.TxID)
		ctx.GetLogger().ErrorOnError(err, "failed to check status")
		fmt.Printf("status is %v\n", status)
		if !pool.IsPending(status) {
			t.Status = StatusNewDone
			return nil, nil
		}
	}
	return self, nil
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
