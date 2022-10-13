package porter

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/coreth/plugin/evm"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/pkg/errors"
)

var _ porter.Txs = (*Txs)(nil)

type Txs struct {
	ExportTxArgs *pchain.ExportTxArgs
	ImportTxArgs *cchain.ImportTxArgs

	exportTx *txs.ExportTx
	importTx *evm.UnsignedImportTx

	exportTxSigBytes [65]byte
	importTxSigBytes [65]byte

	PChainTx *txs.Tx
	CChainTx *evm.Tx
}

func (t *Txs) ExportTxHash() ([]byte, error) {
	exportTx := pchain.ExportTx(t.ExportTxArgs)
	t.exportTx = exportTx

	tx := txs.Tx{
		Unsigned: exportTx,
	}

	unsignedBytes, err := txs.Codec.Marshal(txs.Version, &tx.Unsigned)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func (t *Txs) ImportTxHash() ([]byte, error) {
	exportTxUTXOs := myAvax.UTXOsFromTransferableOutputs(t.PChainTx.ID(), t.exportTx.ExportedOutputs)
	if len(exportTxUTXOs) == 0 {
		return nil, errors.Errorf("no exportTx UTXOs provided.")
	}
	t.ImportTxArgs.AtomicUTXOs = exportTxUTXOs
	importTx := cchain.ImportTx(t.ImportTxArgs)

	fee, err := importTx.GasUsed(true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate atomicTx fee")
	}
	initialOutAmount := importTx.Outs[0].Amount
	adjustOutAmount := initialOutAmount - fee*cchain.BaseFeeGwei
	importTx.Outs[0].Amount = adjustOutAmount
	t.ImportTxArgs.OutAmount = adjustOutAmount

	t.importTx = importTx

	tx := evm.Tx{
		UnsignedAtomicTx: importTx,
	}

	unsignedBytes, err := evm.Codec.Marshal(uint16(0), &tx.UnsignedAtomicTx) // todo: consider config codec version
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func (t *Txs) SetExportTxSig(exportTxSig [65]byte) error {
	signedExportTx, err := pchain.SignedTx(t.exportTx, exportTxSig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.PChainTx = signedExportTx
	t.exportTxSigBytes = exportTxSig
	return nil
}

func (t *Txs) SetImportTxSig(importTxSig [65]byte) error {
	signedImportTx, err := cchain.SignedTx(t.importTx, importTxSig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.CChainTx = signedImportTx
	t.importTxSigBytes = importTxSig
	return nil
}

func (t *Txs) SignedExportTxBytes() ([]byte, error) {
	return t.PChainTx.Bytes(), nil
}

func (t *Txs) SignedImportTxBytes() ([]byte, error) {
	return t.CChainTx.SignedBytes(), nil
}

func (t *Txs) SignedExportTxID() ids.ID {
	return t.PChainTx.ID()
}

func (t *Txs) SignedImportTxID() ids.ID {
	return t.CChainTx.ID()
}
