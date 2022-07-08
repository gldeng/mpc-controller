package avax

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/jinzhu/copier"
)

type MpcUTXO struct {
	UTXOID MpcUTXOID
	Asset  ids.ID
	Out    *MpcTransferOutput
}

type MpcUTXOID struct {
	TxID        ids.ID
	OutputIndex uint32
}

type MpcTransferOutput struct {
	Amt          uint64
	OutputOwners MpcOutputOwners
}

type MpcOutputOwners struct {
	Locktime  uint64
	Threshold uint32
	Addrs     []ids.ShortID
}

func MpcUTXOsFromUTXOs(utxos []*avax.UTXO) []*MpcUTXO {
	var mpcUTXOs []*MpcUTXO
	for _, utxo := range utxos {
		mpcUTXO := MpcUTXOFromUTXO(utxo)
		mpcUTXOs = append(mpcUTXOs, mpcUTXO)
	}
	return mpcUTXOs
}

func MpcUTXOFromUTXO(utxo *avax.UTXO) *MpcUTXO {
	out := utxo.Out.(*secp256k1fx.TransferOutput)

	outputOwners := MpcOutputOwners{}
	copier.Copy(&outputOwners, out.OutputOwners)

	transferOutput := &MpcTransferOutput{
		Amt:          out.Amt,
		OutputOwners: outputOwners,
	}

	mpcUTXO := &MpcUTXO{
		UTXOID: MpcUTXOID{utxo.TxID, utxo.OutputIndex},
		Asset:  utxo.Asset.ID,
		Out:    transferOutput,
	}

	return mpcUTXO
}

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
