package stake

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

const (
	StatusStarted Status = iota
	StatusBuilt

	StatusExportTxSigningPosted
	StatusExportTxSigningDone

	StatusImportTxSigningPosted
	StatusImportTxSigningDone

	StatusExportTxIssued
	StatusExportTxAccepted

	StatusImportTxIssued
	StatusImportTxCommitted
)

type Status int

type StakeTransferTask struct {
	Ctx    context.Context
	Logger logger.Logger

	NonceGiver noncer.Noncer
	Network    chain.NetworkContext

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeAtomicTransferTask]

	Joined *events.RequestStarted

	txs      *Txs
	signReqs []*core.SignRequest

	exportIssueTx   *txissuer.Tx
	exportTxSignRes *core.Result

	importIssueTx   *txissuer.Tx
	importTxSignRes *core.Result

	status Status
}

func (t *StakeTransferTask) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

// todo: function extraction
// todo: enhance resusebility

func (t *StakeTransferTask) do() bool {
	switch t.status {
	case StatusStarted:
		err := t.buildTask()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to build")
			return false
		}
		t.status = StatusBuilt
	case StatusBuilt:
		err := t.MpcClient.Sign(t.Ctx, t.signReqs[0])
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to post ExportTx signing request")
			return false
		}
		t.status = StatusExportTxSigningPosted
	case StatusExportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.signReqs[0].ReqID)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to check ExportTx signing result")
			return false
		}
		if res.Status != core.StatusDone {
			if strings.Contains(string(res.Status), "ERROR") {
				t.Logger.ErrorOnError(errors.New(string(res.Status)), "StakeTransferTask failed to sign ExportTx")
				return false
			}
			t.Logger.Debug("StakeTransferTask hasn't finished ExportTx signing")
			return true
		}
		t.status = StatusExportTxSigningDone
		t.exportTxSignRes = res
	case StatusExportTxSigningDone:
		sig := new(events.Signature).FromHex(t.exportTxSignRes.Result)
		err := t.txs.SetExportTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to set ExportTx signature")
			return false
		}

		signReq, err := t.buildImportTxSignReq(t.txs)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to build ImportTx signing request")
			return false
		}

		t.signReqs[1] = signReq
		err = t.MpcClient.Sign(t.Ctx, t.signReqs[1])
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to post ImportTx signing request")
			return false
		}
		t.status = StatusImportTxSigningPosted
	case StatusImportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.signReqs[1].ReqID)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to check ImportTx signing result")
			return false
		}

		if res.Status != core.StatusDone {
			if strings.Contains(string(res.Status), "ERROR") {
				t.Logger.ErrorOnError(errors.New(string(res.Status)), "StakeTransferTask failed to sign ImportTx")
				return false
			}
			t.Logger.Debug("StakeTransferTask hasn't finished ImportTx signing")
			return true
		}
		t.status = StatusImportTxSigningDone
		t.importTxSignRes = res
	case StatusImportTxSigningDone:
		sig := new(events.Signature).FromHex(t.importTxSignRes.Result)
		err := t.txs.SetImportTxSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to set ImportTx signature")
			return false
		}

		signedBytes, err := t.txs.SignedExportTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to get signed ExportTx bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.signReqs[0].ReqID,
			TxID:  t.txs.ExportTxID(),
			Chain: txissuer.ChainC,
			Bytes: signedBytes,
		}
		t.exportIssueTx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.exportIssueTx)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to issue ExportTx")
			return false
		}
		t.status = StatusExportTxIssued
	case StatusExportTxIssued:
		err := t.TxIssuer.TrackTx(t.Ctx, t.exportIssueTx)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to track ExportTx status")
			return false
		}
		switch t.exportIssueTx.Status {
		case txissuer.StatusFailed:
			t.Logger.Debug(fmt.Sprintf("StakeTransferTask ExporTx failed because of %v", t.exportIssueTx.Reason))
			return false
		case txissuer.StatusAccepted:
			t.Logger.Info("StakeTransferTask ExportTx accepted") // todo: add more context info
			t.status = StatusExportTxAccepted
		}
	case StatusExportTxAccepted:
		signedBytes, err := t.txs.SignedImportTxBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to get signed ImportTx bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.signReqs[1].ReqID,
			TxID:  t.txs.ImportTxID(),
			Chain: txissuer.ChainP,
			Bytes: signedBytes,
		}
		t.importIssueTx = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.importIssueTx)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to issue ImportTx")
			return false // todo: mpc-controller need to join AddDelegator signing process
		}
		t.status = StatusImportTxIssued
	case StatusImportTxIssued:
		err := t.TxIssuer.TrackTx(t.Ctx, t.importIssueTx)
		if err != nil {
			t.Logger.ErrorOnError(err, "StakeTransferTask failed to track ImportTx status")
			return false // todo: mpc-controller need to join AddDelegator signing process
		}

		switch t.importIssueTx.Status {
		case txissuer.StatusFailed:
			t.Logger.Debug(fmt.Sprintf("StakeTransferTask ImportTx failed because of %v", t.importIssueTx.Reason))
			fallthrough // todo: currently mpc-controller need to join signing AddDelegatorTx, consider seperate joining process, using UTXO memo.
		case txissuer.StatusCommitted:
			t.status = StatusImportTxCommitted
			utxos, _ := t.txs.SingedImportTxUTXOs()
			evt := events.StakeAtomicTransferTask{
				StakeTaskBasic: events.StakeTaskBasic{
					ReqNo:   t.txs.ReqNo,
					Nonce:   t.txs.Nonce,
					ReqHash: t.txs.ReqHash,

					DelegateAmt: t.txs.DelegateAmt,
					StartTime:   t.txs.StartTime,
					EndTime:     t.txs.EndTime,
					NodeID:      t.txs.NodeID,

					StakePubKey:   t.Joined.CompressedGenPubKeyHex,
					CChainAddress: t.txs.CChainAddress,
					PChainAddress: t.txs.PChainAddress,

					JoinedPubKeys: t.Joined.CompressedPartiPubKeys,
				},

				ExportTxID: t.exportIssueTx.TxID,
				ImportTxID: t.importIssueTx.TxID,

				UTXOsToStake: utxos,
			}

			t.Dispatcher.Dispatch(&evt) // todo:
			t.Logger.Info("StakeTransferTask finished", []logger.Field{{"StakeAtomicTransferTask", evt}}...)
			return false
		}
	}
	return true
}

// Build task

func (t *StakeTransferTask) buildTask() error {
	stakeReq := t.Joined.JoinedReq.Args.(*storage.StakeRequest)
	txs, err := t.buildTxs(stakeReq, *t.Joined.ReqHash)
	if err != nil {
		return errors.Wrapf(err, "failed to build txs")
	}

	signReq, err := t.buildExportTxSignReq(txs)
	if err != nil {
		return errors.Wrapf(err, "failed to build ExportTx signing request")
	}

	t.txs = txs
	t.signReqs = make([]*core.SignRequest, 2)
	t.signReqs[0] = signReq
	return nil
}

func (t *StakeTransferTask) buildExportTxSignReq(txs *Txs) (*core.SignRequest, error) {
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

func (t *StakeTransferTask) buildImportTxSignReq(txs *Txs) (*core.SignRequest, error) {
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

func (t *StakeTransferTask) buildTxs(stakeReq *storage.StakeRequest, reqHash storage.RequestHash) (*Txs, error) {
	nodeID, _ := ids.ShortFromPrefixedString(stakeReq.NodeID, ids.NodeIDPrefix)
	amountBig := new(big.Int)
	amount, _ := amountBig.SetString(stakeReq.Amount, 10)

	startTime := big.NewInt(stakeReq.StartTime)
	endTIme := big.NewInt(stakeReq.EndTime)

	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(1_000_000_000))
	if !nAVAXAmount.IsUint64() || !startTime.IsUint64() || !endTIme.IsUint64() {
		return nil, errors.New(ErrMsgInvalidUint64)
	}

	cChainAddr, _ := stakeReq.GenPubKey.CChainAddress()
	pChainAddr, _ := stakeReq.GenPubKey.PChainAddress()

	st := Txs{
		ReqNo:         stakeReq.ReqNo,
		Nonce:         t.NonceGiver.GetNonce(stakeReq.ReqNo),
		ReqHash:       reqHash.String(),
		DelegateAmt:   nAVAXAmount.Uint64(),
		StartTime:     startTime.Uint64(),
		EndTime:       endTIme.Uint64(),
		CChainAddress: cChainAddr,
		PChainAddress: pChainAddr,
		NodeID:        ids.NodeID(nodeID),

		BaseFeeGwei: cchain.BaseFeeGwei,
		NetworkID:   t.Network.NetworkID(),
		CChainID:    t.Network.CChainID(),
		Asset:       t.Network.Asset(),
		ImportFee:   t.Network.ImportFee(),
	}

	return &st, nil
}
