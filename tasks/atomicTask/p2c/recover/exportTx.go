package stake

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/pkg/errors"
)

type ExportTx struct {
	Args *pchain.ExportTxArgs

	unsigned *txs.ExportTx
	signed   *txs.Tx
}

func (t *ExportTx) Hash() ([]byte, error) {
	unsigned := pchain.ExportTx(t.Args)
	t.unsigned = unsigned

	tx := txs.Tx{
		Unsigned: unsigned,
	}

	bytes, err := txs.Codec.Marshal(txs.Version, &tx.Unsigned)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(bytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func (t *ExportTx) SetSig(sig [65]byte) error {
	signed, err := pchain.SignedTx(t.unsigned, sig)
	if err != nil {
		return errors.WithStack(err)
	}
	t.signed = signed
	return nil
}

func (t *ExportTx) SignedBytes() ([]byte, error) {
	return t.signed.Bytes(), nil
}

func (t *ExportTx) ID() ids.ID {
	return t.signed.ID()
}

func (t *ExportTx) ExportedOutputs() []*avax.TransferableOutput {
	return t.unsigned.ExportedOutputs
}
