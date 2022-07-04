package avax

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
)

func UTXOsFromTransferableOutputs(txID ids.ID, outputs []*avax.TransferableOutput) []*avax.UTXO {
	var utxos []*avax.UTXO
	for index, output := range outputs {
		utxo := UTXOFromTransferableOutput(txID, output, uint32(index))
		utxos = append(utxos, utxo)
	}

	return utxos
}

func UTXOFromTransferableOutput(txID ids.ID, output *avax.TransferableOutput, outputIndex uint32) *avax.UTXO {
	utxo := avax.UTXO{
		UTXOID: avax.UTXOID{
			TxID:        txID,
			OutputIndex: outputIndex,
		},
		Asset: output.Asset,
		Out:   output.Out,
	}

	return &utxo
}
