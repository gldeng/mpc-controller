package stake

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto/secp256k1r"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/common"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"strconv"
)

const (
	StatusStarted Status = iota
	StatusBuilt

	StatusExportTxSigningPosted
	StatusExportTxSigningDone
	StatusExportTxIssued
	StatusExportTxCommitted

	StatusImportTxSigningPosted
	StatusImportTxSigningDone
	StatusImportTxIssued
	StatusImportTxAccepted
)

const (
	utxoOutputIndexPrincipal utxoOutputIndex = iota
	utxoOutputIndexReward
)

type Status int
type utxoOutputIndex int

type Task struct {
	Ctx    context.Context
	Logger logger.Logger

	Network chain.NetworkContext

	ContractCaller caller.Caller

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.UTXOExported]

	Joined *events.RequestStarted

	UTXOsCache *ristretto.Cache

	exportTx *ExportTx
	importTx *ImportTx

	signReqs    []*core.SignRequest
	sigVerifier *secp256k1r.Verifier

	exportTx_       *txissuer.Tx
	exportTxSignRes *core.Result

	importTx_       *txissuer.Tx
	importTxSignRes *core.Result

	status Status
}

func (t *Task) Do() {
	if t.do() {
		t.Pool.Submit(t.Do)
	}
}

// todo: function extraction

func (t *Task) do() bool {
	switch t.status {
	case StatusStarted:
		err := t.buildTask()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to build task")
			return false
		}
		t.status = StatusBuilt
	case StatusBuilt:
		err := t.MpcClient.Sign(t.Ctx, t.signReqs[0])
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.status = StatusExportTxSigningPosted
		}
	case StatusExportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.signReqs[0].ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.status = StatusExportTxSigningDone
		t.exportTxSignRes = res
	case StatusExportTxSigningDone:
		sig := new(events.Signature).FromHex(t.exportTxSignRes.Result)
		ok, err := t.sigVerifier.VerifySig(bytes.HexToBytes(t.signReqs[0].Hash), *sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to verify signature")
			return false
		}

		if !ok {
			t.Logger.Error("Invalid signature")
			return false
		}

		err = t.exportTx.SetSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to set signature")
			return false
		}

		signedBytes, err := t.exportTx.SignedBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.signReqs[0].ReqID,
			Kind:  events.TxKindPChainExport,
			Bytes: signedBytes,
		}
		t.exportTx_ = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.exportTx_)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.status = StatusExportTxIssued
		}
	case StatusExportTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.exportTx_)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.status = StatusExportTxCommitted
		}
	case StatusExportTxCommitted:
		err := t.MpcClient.Sign(t.Ctx, t.signReqs[1])
		t.Logger.ErrorOnError(err, "Failed to post signing request")
		if err == nil {
			t.status = StatusImportTxSigningPosted
		}
	case StatusImportTxSigningPosted:
		res, err := t.MpcClient.Result(t.Ctx, t.signReqs[1].ReqID)
		t.Logger.ErrorOnError(err, "Failed to check signing result")

		if res.ReqStatus != events.ReqStatusDone {
			t.Logger.Debug("Signing task not done")
			return true
		}
		t.status = StatusImportTxSigningDone
		t.importTxSignRes = res
	case StatusImportTxSigningDone:
		sig := new(events.Signature).FromHex(t.importTxSignRes.Result)
		ok, err := t.sigVerifier.VerifySig(bytes.HexToBytes(t.signReqs[0].Hash), *sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to verify signature")
			return false
		}

		if !ok {
			t.Logger.Error("Invalid signature")
			return false
		}

		err = t.importTx.SetSig(*sig)
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to set signature")
			return false
		}

		signedBytes, err := t.importTx.SignedBytes()
		if err != nil {
			t.Logger.ErrorOnError(err, "Failed to get signed bytes")
			return false
		}

		tx := txissuer.Tx{
			ReqID: t.signReqs[1].ReqID,
			Kind:  events.TxKindPChainImport,
			Bytes: signedBytes,
		}
		t.importTx_ = &tx

		err = t.TxIssuer.IssueTx(t.Ctx, t.importTx_)
		t.Logger.ErrorOnError(err, "Failed to issue tx")
		if err == nil {
			t.status = StatusImportTxIssued
		}
	case StatusImportTxIssued:
		status, err := t.TxIssuer.TrackTx(t.Ctx, t.importTx_)
		t.Logger.ErrorOnError(err, "Failed to track tx")
		if err == nil && status == txissuer.StatusFailed {
			t.Logger.ErrorOnError(err, "Tx failed")
			return false
		}

		if err == nil && status == txissuer.StatusApproved {
			t.status = StatusImportTxAccepted
		}

		utxo := t.exportTx.Args.UTXOs[0]
		evt := events.UTXOExported{
			NativeUTXO:   utxo,
			MpcUTXO:      myAvax.MpcUTXOFromUTXO(utxo),
			ExportedTxID: t.exportTx.ID(),
			ImportedTxID: t.importTx.ID(),
		}

		t.Dispatcher.Dispatch(&evt)
		return false
	}
	return true
}

// Build task

func (t *Task) buildTask() error {
	req := t.Joined.JoinedReq.Args.(*storage.RecoverRequest)
	sigVerifier, err := t.buildSigVerifier(req.GenPubKey)
	if err != nil {
		return errors.WithStack(err)
	}

	exportTx, importTx, err := t.buildTxs(req)
	if err != nil {
		return errors.WithStack(err)
	}

	signReqs, err := t.buildSignReqs(t.Joined.Raw.TxHash.Hex(), exportTx, importTx)
	if err != nil {
		return errors.WithStack(err)
	}

	t.exportTx = exportTx
	t.importTx = importTx
	t.signReqs = signReqs
	t.sigVerifier = sigVerifier
	return nil
}

func (t *Task) buildSigVerifier(signPubKey storage.PubKey) (*secp256k1r.Verifier, error) {
	pChainAddr, err := signPubKey.PChainAddress()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &secp256k1r.Verifier{PChainAddress: pChainAddr}, nil
}

func (t *Task) buildSignReqs(reqHash string, exportTx *ExportTx, importTx *ImportTx) ([]*core.SignRequest, error) {
	exportTxHash, err := exportTx.Hash()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	exportReqPrefix := t.reqIDPrefix(utxoOutputIndex(exportTx.Args.UTXOs[0].OutputIndex), true)
	exportTxSignReq := core.SignRequest{
		ReqID:                  string(exportReqPrefix) + reqHash,
		Kind:                   events.SignKindStakeExport,
		CompressedGenPubKeyHex: t.Joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: t.Joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(exportTxHash),
	}

	importTxHash, err := importTx.Hash()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	importReqPrefix := t.reqIDPrefix(utxoOutputIndex(exportTx.Args.UTXOs[0].OutputIndex), true)
	importTxSignReq := core.SignRequest{
		ReqID:                  string(importReqPrefix) + reqHash,
		Kind:                   events.SignKindStakeImport,
		CompressedGenPubKeyHex: t.Joined.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: t.Joined.CompressedPartiPubKeys,
		Hash:                   bytes.BytesToHex(importTxHash),
	}

	return []*core.SignRequest{&exportTxSignReq, &importTxSignReq}, nil
}

func (t *Task) buildTxs(req *storage.RecoverRequest) (*ExportTx, *ImportTx, error) {
	pChainAddr, _ := req.GenPubKey.PChainAddress()
	treasureAddr, _ := t.treasuryAddress(t.Ctx, utxoOutputIndex(req.OutputIndex))

	utxoID := req.TxID.String() + strconv.Itoa(int(req.OutputIndex))
	val, ok := t.UTXOsCache.Get(utxoID)
	if !ok {
		return nil, nil, errors.Errorf("UTXO(%v) to pay not cached", utxoID)
	}
	utxo := val.(*avax.UTXO)

	amountToExport := utxo.Out.(*secp256k1fx.TransferOutput).Amount()
	if amountToExport < t.Network.ExportFee() {
		return nil, nil, errors.Errorf("insufficient fund: export amount(%v) is less than export fee(%v)", amountToExport, t.Network.ExportFee())
	}
	outAmount := amountToExport - t.Network.ExportFee()

	exportTxArgs := &pchain.ExportTxArgs{
		NetworkID:          t.Network.NetworkID(),
		BlockchainID:       ids.Empty,
		DestinationChainID: t.Network.CChainID(),
		OutAmount:          outAmount,
		To:                 pChainAddr,
		UTXOs:              []*avax.UTXO{utxo},
	}

	importTxArgs := &cchain.ImportTxArgs{
		NetworkID:     t.Network.NetworkID(),
		BlockchainID:  t.Network.CChainID(),
		OutAmount:     outAmount,
		SourceChainID: ids.Empty,
		To:            treasureAddr,
	}

	exportTx := ExportTx{Args: exportTxArgs}
	importTx := ImportTx{Args: importTxArgs, ExportTx: &exportTx}

	return &exportTx, &importTx, nil
}

func (t *Task) reqIDPrefix(outputIndex utxoOutputIndex, export bool) events.ReqIDPrefix {
	var prefix events.ReqIDPrefix
	switch outputIndex {
	case utxoOutputIndexPrincipal:
		if export {
			prefix = events.ReqIDPrefixRecoverPrincipalExport
		} else {
			prefix = events.ReqIDPrefixRecoverPrincipalImport
		}
	case utxoOutputIndexReward:
		if export {
			prefix = events.ReqIDPrefixRecoverRewardExport
		} else {
			prefix = events.ReqIDPrefixRecoverRewardImport
		}
	}
	return prefix
}

func (t *Task) treasuryAddress(ctx context.Context, outputIndex utxoOutputIndex) (addr common.Address, err error) {
	switch outputIndex {
	case utxoOutputIndexPrincipal:
		if addr, err = t.ContractCaller.PrincipalTreasuryAddress(ctx, nil); err != nil {
			return *new(common.Address), errors.WithStack(err)
		}

	case utxoOutputIndexReward:
		if addr, err = t.ContractCaller.RewardTreasuryAddress(ctx, nil); err != nil {
			return *new(common.Address), errors.WithStack(err)
		}
	}
	return
}
