package p2c

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	myAvax "github.com/avalido/mpc-controller/utils/txs/avax"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type TxBuilder struct {
	net *core.NetworkContext
}

func NewTxBuilder(net *core.NetworkContext) *TxBuilder {
	return &TxBuilder{net: net}
}

func (t *TxBuilder) ExportFromPChain(utxos []*avax.UTXO) (*txs.ExportTx, error) {
	inputs := myAvax.TransferableInputsrFromUTXOs(utxos) // The inputs to this transaction
	avax.SortTransferableInputs(inputs)
	utxoFirst := utxos[0]
	outFirst := utxoFirst.Out.(*secp256k1fx.TransferOutput)
	amount := myAvax.TotalAmount(utxos)
	feeAmount := uint64(1000000)
	if amount < feeAmount {
		return nil, errors.New("inputs are insufficient to pay transaction fee")
	}
	netAmount := amount - feeAmount
	outputs := []*avax.TransferableOutput{{ // Outputs that are exported to the destination chain
		Asset: utxoFirst.Asset,
		Out: &secp256k1fx.TransferOutput{
			Amt:          netAmount,
			OutputOwners: outFirst.OutputOwners,
		},
	}}
	utx := &txs.ExportTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    t.net.NetworkID(),
			BlockchainID: ids.Empty,
			Ins:          inputs,
		}},
		DestinationChain: t.net.CChainID(),
		ExportedOutputs:  outputs,
	}
	return utx, nil
}

func (t *TxBuilder) ImportIntoCChain(to common.Address, signedExportTx *txs.Tx, memo []byte) (*evm.UnsignedImportTx, error) {
	utxos := myAvax.UTXOsFromTransferableOutputs(signedExportTx.ID(), signedExportTx.Unsigned.(*txs.ExportTx).ExportedOutputs)
	inputs := myAvax.TransferableInputsrFromUTXOs(utxos)
	avax.SortTransferableInputs(inputs)

	feeAmount := uint64(1000000)

	totalAmount := uint64(0)
	for _, utxo := range utxos {
		totalAmount += utxo.Out.(*secp256k1fx.TransferOutput).Amt
	}
	utx := &evm.UnsignedImportTx{
		NetworkID:      t.net.NetworkID(),
		BlockchainID:   t.net.CChainID(),
		SourceChain:    ids.Empty,
		ImportedInputs: inputs, // Inputs that consume UTXOs produced on the chain
		Outs: []evm.EVMOutput{
			{
				Address: to,
				Amount:  totalAmount - feeAmount,
				AssetID: utxos[0].Asset.ID,
			},
		},
	}
	return utx, nil
}
