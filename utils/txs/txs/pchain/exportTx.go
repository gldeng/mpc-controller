package pchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	myAvax "github.com/avalido/mpc-controller/utils/txs/avax"
)

type ExportTxArgs struct {
	NetworkID          uint32 // ID of the network this chain lives on
	BlockchainID       ids.ID // ID of the chain on which this transaction exists (prevents replay attacks)
	DestinationChainID ids.ID // Which chain to send the funds to
	OutAmount          uint64
	To                 ids.ShortID
	UTXOs              []*avax.UTXO // UTXOs to spend
}

func ExportTx(args *ExportTxArgs) *txs.ExportTx {
	inputs := myAvax.TransferableInputsrFromUTXOs(args.UTXOs) // The inputs to this transaction
	outputs := []*avax.TransferableOutput{{                   // Outputs that are exported to the destination chain
		Asset: args.UTXOs[0].Asset,
		Out: &secp256k1fx.TransferOutput{
			Amt: args.OutAmount,
			OutputOwners: secp256k1fx.OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs:     []ids.ShortID{args.To},
			},
		},
	}}
	utx := &txs.ExportTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    args.NetworkID,
			BlockchainID: args.BlockchainID,
			Ins:          inputs,
		}},
		DestinationChain: args.DestinationChainID,
		ExportedOutputs:  outputs,
	}
	return utx
}
