package cchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/coreth/plugin/evm"
	myAvax "github.com/avalido/mpc-controller/utils/txs/avax"
	"github.com/ethereum/go-ethereum/common"
)

type ImportTxArgs struct {
	NetworkID     uint32 // ID of the network on which this tx was issued
	BlockchainID  ids.ID // ID of this blockchain.
	SourceChainID ids.ID // Which chain to consume the funds from
	OutAmount     uint64
	To            common.Address // Address of recipient
	AtomicUTXOs   []*avax.UTXO   // UTXOs to spend
}

func ImportTx(args *ImportTxArgs) *evm.UnsignedImportTx {
	utx := &evm.UnsignedImportTx{
		NetworkID:      args.NetworkID,
		BlockchainID:   args.BlockchainID,
		SourceChain:    args.SourceChainID,
		ImportedInputs: myAvax.TransferableInputsrFromUTXOs(args.AtomicUTXOs), // Inputs that consume UTXOs produced on the chain
		Outs: []evm.EVMOutput{
			{
				Address: args.To,
				Amount:  args.OutAmount,
				AssetID: args.AtomicUTXOs[0].AssetID(),
			},
		},
	}
	return utx
}
