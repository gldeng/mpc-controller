package avax

import (
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

func TransferableInputsrFromUTXOs(utxos []*avax.UTXO) []*avax.TransferableInput {
	inputs := []*avax.TransferableInput{}

	for _, utxo := range utxos {
		input := TransferableInputFromUTXO(utxo)
		inputs = append(inputs, input)
	}
	return inputs
}

func TransferableInputFromUTXO(utxo *avax.UTXO) *avax.TransferableInput {
	return &avax.TransferableInput{
		UTXOID: utxo.UTXOID,
		Asset:  utxo.Asset,
		In: &secp256k1fx.TransferInput{
			Amt: utxo.Out.(*secp256k1fx.TransferOutput).Amt,
			Input: secp256k1fx.Input{
				SigIndices: []uint32{0},
			},
		},
	}
}
