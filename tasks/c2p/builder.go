package c2p

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/storage"
	"github.com/pkg/errors"
)

type TxBuilder struct {
	net *core.NetworkContext
}

func NewTxBuilder(net *core.NetworkContext) *TxBuilder {
	return &TxBuilder{net: net}
}

func (t *TxBuilder) ExportFromCChain(pubKey storage.PubKey, amount, nonce uint64) (*evm.UnsignedExportTx, error) {
	exportAmt := amount + t.net.ImportFee()
	cChaiAddress, err := pubKey.CChainAddress()
	if err != nil {
		return nil, err
	}
	pChaiAddress, err := pubKey.PChainAddress()
	if err != nil {
		return nil, err
	}
	asset := t.net.Asset()
	input := evm.EVMInput{
		Address: cChaiAddress,
		Amount:  exportAmt,
		AssetID: (&asset).AssetID(),
		Nonce:   nonce,
	}
	var outs []*avax.TransferableOutput
	outs = append(outs, &avax.TransferableOutput{
		Asset: t.net.Asset(),
		Out: &secp256k1fx.TransferOutput{
			Amt: exportAmt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					pChaiAddress,
				},
			},
		},
	})

	tx := &evm.UnsignedExportTx{
		NetworkID:        t.net.NetworkID(),
		BlockchainID:     t.net.CChainID(),
		DestinationChain: ids.Empty,
		Ins: []evm.EVMInput{
			input,
		},
		ExportedOutputs: outs,
	}

	gas, err := tx.GasUsed(true)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	exportFee := gas * t.net.BaseFeeGwei()
	tx.Ins[0].Amount += exportFee
	return tx, nil
}

func (t *TxBuilder) ImportIntoPChain(pubKey storage.PubKey, signedExportTx *evm.Tx) (*txs.ImportTx, error) {
	exportTx, ok := signedExportTx.UnsignedAtomicTx.(*evm.UnsignedExportTx)
	if !ok {
		return nil, errors.New("not a valid ExportTx")
	}
	index := uint32(0)
	amt := exportTx.ExportedOutputs[index].Out.Amount()
	utxo, err := t.TransferTo(pubKey, amt-t.net.ImportFee())
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare utxo")
	}
	tx := &txs.ImportTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    t.net.NetworkID(),
			BlockchainID: ids.Empty,
			Outs: []*avax.TransferableOutput{
				utxo,
			},
		}},
		SourceChain: t.net.CChainID(),
		ImportedInputs: []*avax.TransferableInput{{
			UTXOID: avax.UTXOID{
				TxID:        exportTx.ID(),
				OutputIndex: index,
			},
			Asset: t.net.Asset(),
			In: &secp256k1fx.TransferInput{
				Amt: amt,
				Input: secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
		}},
	}
	return tx, nil
}

func (t *TxBuilder) TransferTo(pubKey storage.PubKey, amt uint64) (*avax.TransferableOutput, error) {
	pChainAddress, err := pubKey.PChainAddress()
	if err != nil {
		return nil, err
	}
	return &avax.TransferableOutput{
		Asset: t.net.Asset(),
		Out: &secp256k1fx.TransferOutput{
			Amt: amt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					pChainAddress,
				},
			},
		},
	}, nil
}
