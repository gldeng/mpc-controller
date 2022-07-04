package pchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
)

type Args struct {
	NetworkID          uint32
	BlockchainID       ids.ID
	DestinationChainID ids.ID
	AssetID            ids.ID
	Amount             uint64
	To                 ids.ShortID
	UTXOs              []*avax.UTXO
}

func UnsignedExportTx(args *Args) *platformvm.UnsignedExportTx {
	utx := &platformvm.UnsignedExportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    args.NetworkID,
			BlockchainID: args.BlockchainID,
			Ins:          myAvax.TransferableInputsrFromUTXOs(args.UTXOs),
		}},
		DestinationChain: args.DestinationChainID,
		ExportedOutputs: []*avax.TransferableOutput{{
			Asset: avax.Asset{ID: args.AssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: args.Amount,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{args.To},
				},
			},
		}},
	}
	return utx
}
