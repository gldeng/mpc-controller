package c2p

import (
	"context"
	"encoding/hex"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/utils/bytes"
	"math/big"
)

var (
	_ pool.Task = (*ExportFromCChain)(nil)
)

type ExportFromCChain struct {
	Status      Status
	Id          string
	Amount      big.Int
	Quorum      QuorumInfo
	Tx          *evm.UnsignedExportTx
	TxHash      []byte
	TxCred      *secp256k1fx.Credential
	TxID        *ids.ID
	SignRequest *core.SignRequest
	//Txs             *Txs
	//SignReqs        []*core.SignRequest
	//ExportIssueTx   *txissuer.Tx
	//ExportTxStatus  pool.Status
	//ExportTxSignRes *core.Result
	//
	//ImportIssueTx   *txissuer.Tx
	//ImportTxStatus  pool.Status
	//ImportTxSignRes *core.Result
}

func (t *ExportFromCChain) IsDone() bool {
	return t.Status == StatusNewDone
}

func NewExportFromCChain(id string, amount big.Int, quorum QuorumInfo) (*ExportFromCChain, error) {
	return &ExportFromCChain{
		Status:      StatusStarted,
		Id:          id,
		Amount:      amount,
		Quorum:      quorum,
		Tx:          nil,
		TxHash:      nil,
		TxCred:      nil,
		SignRequest: nil,
	}, nil
}

// SignRequestSent

func (t *ExportFromCChain) Next(ctx pool.TaskContext) ([]pool.Task, error) {
	self := []pool.Task{t}
	switch t.Status {
	case StatusStarted:
		builder := NewTxBuilder(ctx.GetNetwork())
		nonce, _ := ctx.NonceAt(t.Quorum.CChainAddress())
		amount, _ := ToGwei(&t.Amount)
		tx, _ := builder.ExportFromCChain(t.Quorum.PubKey, amount, nonce)
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
		txCred, _ := ValidateAndGetCred(t.TxHash, *new(events.Signature).FromHex(res.Result), t.Quorum.PChainAddress())
		t.TxCred = txCred
		signed, _ := t.SignedTx()
		txId, _ := ctx.IssuePChainTx(signed.SignedBytes())
		t.TxID = &txId
		t.Status = StatusNewTxSent
	case StatusNewTxSent:
		status, _ := ctx.CheckCChainTx(*t.TxID)
		if !pool.IsPending(status) {
			t.Status = StatusNewDone
			return nil, nil
		}
	}
	return self, nil
}

func (t *ExportFromCChain) SignedTx() (*evm.Tx, error) {
	return PackSignedExportTx(t.Tx, t.TxCred)
}

//func (t *ExportFromCChain) buildTask(ctx pool.TaskContext) error {
//	txs, err := t.buildTxs(ctx)
//	if err != nil {
//		return errors.Wrapf(err, "failed to build txs")
//	}
//
//	signReq, err := t.buildExportTxSignReq(txs)
//	if err != nil {
//		return errors.Wrapf(err, "failed to build ExportTx signing request")
//	}
//
//	t.Txs = txs
//	t.SignReqs = make([]*core.SignRequest, 2)
//	t.SignReqs[0] = signReq
//	return nil
//}

//func (t *ExportFromCChain) buildExportTxSignReq(txs *Txs) (*core.SignRequest, error) {
//	exportTxHash, err := txs.ExportTxHash()
//	if err != nil {
//		return nil, errors.Wrapf(err, "failed to get ExportTx hash")
//	}
//
//	return t.buildSignReq(t.Id+"/export", exportTxHash)
//}
//
//func (t *ExportFromCChain) buildImportTxSignReq(txs *Txs) (*core.SignRequest, error) {
//	importTxHash, err := txs.ImportTxHash()
//	if err != nil {
//		return nil, errors.Wrapf(err, "failed to get ImportTx hash")
//	}
//
//	return t.buildSignReq(t.Id+"/import", importTxHash)
//}

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

//func (t *ExportFromCChain) buildTxs(ctx pool.TaskContext) (*Txs, error) {
//
//	nAVAXAmount := new(big.Int).Div(&t.Amount, big.NewInt(1_000_000_000))
//	if !nAVAXAmount.IsUint64() {
//		return nil, errors.New(ErrMsgInvalidUint64)
//	}
//
//	cChainAddr, _ := (storage.PubKey(t.Quorum.PubKey)).CChainAddress()
//	pChainAddr, _ := (storage.PubKey(t.Quorum.PubKey)).PChainAddress()
//	nonce, _ := ctx.NonceAt(cChainAddr)
//
//	st := Txs{
//		Nonce:         nonce,
//		DelegateAmt:   nAVAXAmount.Uint64(),
//		CChainAddress: cChainAddr,
//		PChainAddress: pChainAddr,
//
//		BaseFeeGwei: cchain.BaseFeeGwei,
//		NetworkID:   ctx.GetNetwork().NetworkID(),
//		CChainID:    ctx.GetNetwork().CChainID(),
//		Asset:       ctx.GetNetwork().Asset(),
//		ImportFee:   ctx.GetNetwork().ImportFee(),
//	}
//
//	return &st, nil
//}
