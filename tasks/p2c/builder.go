package p2c

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/core"
	myAvax "github.com/avalido/mpc-controller/utils/txs/avax"
)

type TxBuilder struct {
	net *core.NetworkContext
}

func NewTxBuilder(net *core.NetworkContext) *TxBuilder {
	return &TxBuilder{net: net}
}

func (t *TxBuilder) ExportFromPChain(utxo avax.UTXO) (*txs.ExportTx, error) {
	inputs := myAvax.TransferableInputsrFromUTXOs([]*avax.UTXO{&utxo}) // The inputs to this transaction
	out := utxo.Out.(*secp256k1fx.TransferOutput)
	feeAmount := uint64(1000000)
	netAmount := out.Amount() - feeAmount
	outputs := []*avax.TransferableOutput{{ // Outputs that are exported to the destination chain
		Asset: utxo.Asset,
		Out: &secp256k1fx.TransferOutput{
			Amt:          netAmount,
			OutputOwners: out.OutputOwners,
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
