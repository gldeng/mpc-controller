package atomicTask

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/ethereum/go-ethereum/common"
)

var _ AtomicTx = (*StakeAtomicTx)(nil)

type StakeAtomicTx struct {
	ReqID string

	ReqNo         uint64
	Nonce         uint64
	ReqHash       string
	DelegateAmt   uint64
	StartTime     uint64
	EndTime       uint64
	CChainAddress common.Address
	PChainAddress ids.ShortID
	NodeID        ids.NodeID
	BaseFeeGwei   uint64

	NetWorkID uint32
	CChainID  ids.ID
	AssetID   ids.ID
	ImportFee uint64

	exportTx *evm.UnsignedExportTx
}

func (t *StakeAtomicTx) RequestID() string {
	return t.ReqID
}

func (t *StakeAtomicTx) SourceChain() SourceChain {
	return SourceChainCChain
}

func (t *StakeAtomicTx) ExportTxHash() ([]byte, error) {
	exportTx, err := t.buildUnsignedExportTx()
	if err != nil {
		return nil, err
	}
	tx := evm.Tx{
		UnsignedAtomicTx: exportTx,
	}
	unsignedBytes, err := evm.Codec.Marshal(evmCodecVersion, &tx.UnsignedAtomicTx)
	if err != nil {
		return nil, err
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *StakeAtomicTx) SetExportTxSig(sig [SigLength]byte) error {
	return nil
}

func (t *StakeAtomicTx) SignedExportTxBytes() ([]byte, error) {
	return nil, nil
}

func (t *StakeAtomicTx) SetExportTxID(id ids.ID) {

}

func (t *StakeAtomicTx) ImportTxHash() ([]byte, error) {
	return nil, nil
}

func (t *StakeAtomicTx) SetImportTxSig(sig [SigLength]byte) error {
	return nil
}

func (t *StakeAtomicTx) SignedImportTxBytes() ([]byte, error) {
	return nil, nil
}

func (t *StakeAtomicTx) SetImportTxID(id ids.ID) {

}

func (t *StakeAtomicTx) String() string {
	return ""
}

// ---
func (t *StakeAtomicTx) buildUnsignedExportTx() (*evm.UnsignedExportTx, error) {
	if t.exportTx != nil {
		return t.exportTx, nil
	}
	exportAmt := t.DelegateAmt + t.ImportFee
	input := evm.EVMInput{
		Address: t.CChainAddress,
		Amount:  exportAmt,
		AssetID: t.AssetID,
		Nonce:   t.Nonce,
	}
	var outs []*avax.TransferableOutput
	outs = append(outs, &avax.TransferableOutput{
		Asset: avax.Asset{ID: t.AssetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: exportAmt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					t.PChainAddress,
				},
			},
		},
	})

	tx := &evm.UnsignedExportTx{
		NetworkID:        t.NetWorkID,
		BlockchainID:     t.CChainID,
		DestinationChain: ids.Empty,
		Ins: []evm.EVMInput{
			input,
		},
		ExportedOutputs: outs,
	}

	gas, err := tx.GasUsed(true)
	if err != nil {
		return nil, err
	}
	exportFee := gas * t.BaseFeeGwei
	tx.Ins[0].Amount += exportFee

	t.exportTx = tx
	return t.exportTx, nil
}
