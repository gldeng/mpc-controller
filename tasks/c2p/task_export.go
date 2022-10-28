package c2p

import (
	"context"
	"encoding/hex"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/utils/bytes"
	"math/big"
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
	self := []core.Task{t}
	switch t.Status {
	case StatusInit:
		builder := NewTxBuilder(ctx.GetNetwork())
		nonce, err := ctx.NonceAt(t.Quorum.CChainAddress())
		ctx.GetLogger().ErrorOnError(err, "failed to get nonce")
		amount, err := ToGwei(&t.Amount)
		ctx.GetLogger().ErrorOnError(err, "failed to convert amount")
		tx, err := builder.ExportFromCChain(t.Quorum.PubKey, amount, nonce)
		ctx.GetLogger().ErrorOnError(err, "failed to build export tx")
		t.Tx = tx
		txHash, err := ExportTxHash(tx)
		t.TxHash = txHash
		ctx.GetLogger().ErrorOnError(err, "failed to get export tx hash")
		req, err := t.buildSignReq(t.Id+"/export", txHash)
		t.SignRequest = req
		ctx.GetLogger().ErrorOnError(err, "failed create sign request")
		err = ctx.GetMpcClient().Sign(context.Background(), req)
		ctx.GetLogger().ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusSignReqSent
		}
	case StatusSignReqSent:
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
		_, err = ctx.IssueCChainTx(signed.SignedBytes())
		ctx.GetLogger().ErrorOnError(err, "failed to issue tx")
		txId := signed.ID()
		t.TxID = &txId
		t.Status = StatusTxSent
	case StatusTxSent:
		status, err := ctx.CheckCChainTx(*t.TxID)
		ctx.GetLogger().ErrorOnError(err, "failed to check status")
		if !core.IsPending(status) {
			t.Status = StatusDone
			return nil, nil
		}
	}
	return self, nil
}

func (t *ExportFromCChain) SignedTx() (*evm.Tx, error) {
	return PackSignedExportTx(t.Tx, t.TxCred)
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
