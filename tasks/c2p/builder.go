package c2p

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/storage"
	"github.com/pkg/errors"
)

type TxBuilder struct {
	net *chain.NetworkContext
}

func NewTxBuilder(net *chain.NetworkContext) *TxBuilder {
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
