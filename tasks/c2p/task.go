package c2p

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/tasks/atomicTask/c2pChain/stake"
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
)

type Status int

type TransferC2P struct {
	Status          Status
	Joined          events.RequestStarted
	Txs             *stake.Txs
	SignReqs        []*core.SignRequest
	ExportIssueTx   *txissuer.Tx
	ExportTxSignRes *core.Result

	ImportIssueTx   *txissuer.Tx
	ImportTxSignRes *core.Result
}

func (t *TransferC2P) Next(ctx *pool.TaskContext) ([]pool.Task, error) {
	self := []pool.Task{t}
	switch t.Status {
	case StatusStarted:
		err := t.buildTask(ctx)
		if err != nil {
			ctx.Logger.ErrorOnError(err, "Failed to build task")
			return nil, err
		}
		t.Status = StatusBuilt
	case StatusBuilt:
		err := ctx.MpcClient.Sign(context.Background(), t.SignReqs[0])
		ctx.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusExportTxSigningPosted
		}
	case StatusExportTxSigningPosted:
		res, err := ctx.MpcClient.Result(context.Background(), t.SignReqs[0].ReqID)
		ctx.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.Status != core.StatusDone {
			ctx.Logger.Debug("Signing task not done")
			return self, nil
		}
		t.Status = StatusExportTxSigningDone
		t.ExportTxSignRes = res
	case StatusExportTxSigningDone:
		sig := new(events.Signature).FromHex(t.ExportTxSignRes.Result)
		err := t.Txs.SetExportTxSig(*sig)
		if err != nil {
			ctx.Logger.ErrorOnError(err, "Failed to set signature")
			return nil, nil
		}

		signedBytes, err := t.Txs.SignedExportTxBytes()
		if err != nil {
			ctx.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return nil, nil
		}

		tx := txissuer.Tx{
			ReqID: t.SignReqs[0].ReqID,
			TxID:  t.Txs.ExportTxID(),
			Chain: txissuer.ChainC,
			Bytes: signedBytes,
		}
		t.ExportIssueTx = &tx

		err = ctx.TxIssuer.IssueTx(context.Background(), t.ExportIssueTx)
		if err == nil {
			t.Status = StatusExportTxIssued
		}
	case StatusExportTxIssued:
		err := ctx.TxIssuer.TrackTx(context.Background(), t.ExportIssueTx)
		if err == nil && t.ExportIssueTx.Status == txissuer.StatusFailed {
			t.Status = StatusExportTxFailed
		}

		if err == nil && t.ExportIssueTx.Status == txissuer.StatusApproved {
			t.Status = StatusExportTxApproved
		}
	case StatusExportTxFailed:
		fallthrough
	case StatusExportTxApproved:
		signReq, err := t.buildImportTxSignReq(t.Txs)
		if err != nil {
			ctx.Logger.ErrorOnError(err, "failed to build ImportTx signing request")
			return nil, nil
		}

		t.SignReqs[1] = signReq

		err = ctx.MpcClient.Sign(context.Background(), t.SignReqs[1])
		ctx.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.Status = StatusImportTxSigningPosted
		}
	case StatusImportTxSigningPosted:
		res, err := ctx.MpcClient.Result(context.Background(), t.SignReqs[1].ReqID)
		ctx.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.Status != core.StatusDone {
			ctx.Logger.Debug("Signing task not done")
			return self, nil
		}
		t.Status = StatusImportTxSigningDone
		t.ImportTxSignRes = res
	case StatusImportTxSigningDone:
		sig := new(events.Signature).FromHex(t.ImportTxSignRes.Result)
		err := t.Txs.SetImportTxSig(*sig)
		if err != nil {
			ctx.Logger.ErrorOnError(err, "Failed to set signature")
			return nil, nil
		}

		signedBytes, err := t.Txs.SignedImportTxBytes()
		if err != nil {
			ctx.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return nil, nil
		}

		tx := txissuer.Tx{
			ReqID: t.SignReqs[1].ReqID,
			TxID:  t.Txs.ImportTxID(),
			Chain: txissuer.ChainP,
			Bytes: signedBytes,
		}
		t.ImportIssueTx = &tx

		err = ctx.TxIssuer.IssueTx(context.Background(), t.ImportIssueTx)
		if err == nil {
			t.Status = StatusImportTxIssued
		}
	case StatusImportTxIssued:
		err := ctx.TxIssuer.TrackTx(context.Background(), t.ImportIssueTx)
		if err == nil && t.ImportIssueTx.Status == txissuer.StatusFailed {
			t.Status = StatusImportTxFailed
		}

		if err == nil && t.ImportIssueTx.Status == txissuer.StatusApproved {
			t.Status = StatusImportTxApproved
		}
	case StatusImportTxFailed:
		fallthrough
	case StatusImportTxApproved:
		utxos, _ := t.Txs.SingedImportTxUTXOs()

		evt := events.StakeAtomicTaskHandled{
			StakeTaskBasic: events.StakeTaskBasic{
				ReqNo:   t.Txs.ReqNo,
				Nonce:   t.Txs.Nonce,
				ReqHash: t.Txs.ReqHash,

				DelegateAmt: t.Txs.DelegateAmt,
				StartTime:   t.Txs.StartTime,
				EndTime:     t.Txs.EndTime,
				NodeID:      t.Txs.NodeID,

				StakePubKey:   t.Joined.CompressedGenPubKeyHex,
				CChainAddress: t.Txs.CChainAddress,
				PChainAddress: t.Txs.PChainAddress,

				JoinedPubKeys: t.Joined.CompressedPartiPubKeys,
			},

			ExportTxID: t.ExportIssueTx.TxID,
			ImportTxID: t.ImportIssueTx.TxID,

			UTXOsToStake: utxos,
		}

		ctx.Dispatcher.Dispatch(&evt)
		ctx.Logger.Info("Stake atomic task handled", []logger.Field{{"stakeAtomicTaskHandled", evt}}...)
		return nil, nil
	}
	return self, nil
}

func (t *TransferC2P) buildTask(ctx *pool.TaskContext) error {
	stakeReq := t.Joined.JoinedReq.Args.(*storage.StakeRequest)
	txs, err := t.buildTxs(ctx, stakeReq, *t.Joined.ReqHash)
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

func (t *TransferC2P) buildExportTxSignReq(txs *stake.Txs) (*core.SignRequest, error) {
	exportTxHash, err := txs.ExportTxHash()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ExportTx hash")
	}

	exportTxSignReq := core.SignRequest{
		ReqID:                  string(events.ReqIDPrefixStakeExport) + fmt.Sprintf("%v", txs.ReqNo) + "-" + txs.ReqHash,
		CompressedGenPubKeyHex: t.Joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: t.Joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(exportTxHash),
	}

	return &exportTxSignReq, nil
}

func (t *TransferC2P) buildImportTxSignReq(txs *stake.Txs) (*core.SignRequest, error) {
	importTxHash, err := txs.ImportTxHash()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ImportTx hash")
	}

	importTxSignReq := core.SignRequest{
		ReqID:                  string(events.ReqIDPrefixStakeImport) + fmt.Sprintf("%v", txs.ReqNo) + "-" + txs.ReqHash,
		CompressedGenPubKeyHex: t.Joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: t.Joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(importTxHash),
	}

	return &importTxSignReq, nil
}

func (t *TransferC2P) buildTxs(ctx *pool.TaskContext, stakeReq *storage.StakeRequest, reqHash storage.RequestHash) (*stake.Txs, error) {
	nodeID, _ := ids.ShortFromPrefixedString(stakeReq.NodeID, ids.NodeIDPrefix)
	amountBig := new(big.Int)
	amount, _ := amountBig.SetString(stakeReq.Amount, 10)

	startTime := big.NewInt(stakeReq.StartTime)
	endTIme := big.NewInt(stakeReq.EndTime)

	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !startTime.IsUint64() || !endTIme.IsUint64() {
		return nil, errors.New(stake.ErrMsgInvalidUint64)
	}

	cChainAddr, _ := stakeReq.GenPubKey.CChainAddress()
	pChainAddr, _ := stakeReq.GenPubKey.PChainAddress()

	st := stake.Txs{
		ReqNo:         stakeReq.ReqNo,
		Nonce:         ctx.NonceGiver.GetNonce(stakeReq.ReqNo),
		ReqHash:       reqHash.String(),
		DelegateAmt:   nAVAXAmount.Uint64(),
		StartTime:     startTime.Uint64(),
		EndTime:       endTIme.Uint64(),
		CChainAddress: cChainAddr,
		PChainAddress: pChainAddr,
		NodeID:        ids.NodeID(nodeID),

		BaseFeeGwei: cchain.BaseFeeGwei,
		NetworkID:   ctx.Network.NetworkID(),
		CChainID:    ctx.Network.CChainID(),
		Asset:       ctx.Network.Asset(),
		ImportFee:   ctx.Network.ImportFee(),
	}

	return &st, nil
}
