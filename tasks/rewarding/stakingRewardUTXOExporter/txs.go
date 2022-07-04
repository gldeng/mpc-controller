package stakingRewardUTXOExporter

import (
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/pkg/errors"
)

var _ porter.Txs = (*Txs)(nil)

type Txs struct {
	UnsignedExportTxArgs *pchain.Args
	UnsignedImportTx     *cchain.Args

	unsignedExportTx *platformvm.UnsignedExportTx
	unsignedImportTx *evm.UnsignedImportTx

	signedExportTx *platformvm.Tx
	signedImportTx *evm.Tx
}

func (t *Txs) ExportTxHash() ([]byte, error) {
	exportTx := pchain.UnsignedExportTx(t.UnsignedExportTxArgs)
	t.unsignedExportTx = exportTx

	tx := platformvm.Tx{
		UnsignedTx: exportTx,
	}

	unsignedBytes, err := platformvm.Codec.Marshal(platformvm.CodecVersion, &tx.UnsignedTx) // todo: consider config codec version
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
	importTx := cchain.UnsignedImportTx(t.UnsignedImportTx)
	t.unsignedImportTx = importTx

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
	signedExportedTx, err := pchain.SignedTx(t.unsignedExportTx, exportTxSig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.signedExportTx = signedExportedTx
	return nil
}

func (t *Txs) SetImportTxSig(importTxSig [65]byte) error {
	signedImportedTx, err := cchain.SignedTx(t.unsignedImportTx, importTxSig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.signedImportTx = signedImportedTx
	return nil
}

func (t *Txs) SignedExportTxBytes() ([]byte, error) {
	return t.signedExportTx.Bytes(), nil
}

func (t *Txs) SignedImportTxBytes() ([]byte, error) {
	return t.signedImportTx.Bytes(), nil
}
