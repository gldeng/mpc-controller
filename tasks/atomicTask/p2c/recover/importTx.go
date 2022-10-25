package stake

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/coreth/plugin/evm"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/pkg/errors"
)

type ImportTx struct {
	Args     *cchain.ImportTxArgs
	ExportTx *ExportTx

	unsigned *evm.UnsignedImportTx
	signed   *evm.Tx
}

func (t *ImportTx) Hash() ([]byte, error) {
	utxos := myAvax.UTXOsFromTransferableOutputs(t.ExportTx.ID(), t.ExportTx.ExportedOutputs())
	if len(utxos) == 0 {
		return nil, errors.New("no UTXO to spend")
	}
	t.Args.AtomicUTXOs = utxos
	unsigned := cchain.ImportTx(t.Args)

	fee, err := unsigned.GasUsed(true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate gas fee")
	}
	initialOutAmount := unsigned.Outs[0].Amount
	adjustOutAmount := initialOutAmount - fee*cchain.BaseFeeGwei
	unsigned.Outs[0].Amount = adjustOutAmount
	t.Args.OutAmount = adjustOutAmount

	t.unsigned = unsigned

	tx := evm.Tx{
		UnsignedAtomicTx: unsigned,
	}

	bytes, err := evm.Codec.Marshal(txs.Version, &tx.UnsignedAtomicTx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(bytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func (t *ImportTx) SetSig(sig [65]byte) error {
	signed, err := cchain.SignedTx(t.unsigned, sig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.signed = signed
	return nil
}

func (t *ImportTx) SignedBytes() ([]byte, error) {
	return t.signed.SignedBytes(), nil
}

func (t *ImportTx) ID() ids.ID {
	return t.signed.ID()
}
