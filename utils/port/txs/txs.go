package txs

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
)

var _ porter.Txs = (*Txs)(nil)

// todo: to improve

type Args struct {
	NetworkID uint32
	AssetID   ids.ID

	ChainID     ids.ID         // chain to import from
	To          common.Address // Address of recipient
	BaseFee     *big.Int       // fee to use post-AP3
	AtomicUTXOs []*avax.UTXO   // UTXOs to spend

	BlockchainID       ids.ID
	DestinationChainID ids.ID
	Amount             uint64
	//To                 ids.ShortID
	Ins []*avax.TransferableInput
}

type Txs struct {
	Args *Args

	unsignedExportTx *platformvm.UnsignedExportTx
	unsignedImportTx *evm.UnsignedImportTx

	exportTxSig [65]byte
	importTxSig [65]byte
}

func (t *Txs) ExportTxHash() ([]byte, error) {
	exportTx := pchain.UnsignedExportTx(nil) // todo: nil
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
	importTx := cchain.UnsignedImportTx(nil) // todo: nil
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
	t.exportTxSig = exportTxSig
	return nil
}

func (t *Txs) SetImportTxSig(importTxSig [65]byte) error {
	t.importTxSig = importTxSig
	return nil
}

func (t *Txs) SignedExportTxBytes() ([]byte, error) {
	signedExportTx, err := pchain.SignedTx(t.unsignedExportTx, t.exportTxSig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return signedExportTx.Bytes(), nil
}

func (t *Txs) SignedImportTxBytes() ([]byte, error) {
	signedExportTx, err := cchain.SignedTx(t.unsignedImportTx, t.importTxSig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return signedExportTx.Bytes(), nil
}
