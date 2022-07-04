package cchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/coreth/plugin/evm"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/ethereum/go-ethereum/common"
)

type Args struct {
	NetworkID     uint32         // ID of the network on which this tx was issued
	BlockchainID  ids.ID         // ID of this blockchain.
	SourceChainID ids.ID         // Which chain to consume the funds from
	To            common.Address // Address of recipient
	//BaseFee      *big.Int       // fee to use post-AP3 // todo: consider this kind of fee
	AtomicUTXOs []*avax.UTXO // UTXOs to spend
}

func UnsignedImportTx(args *Args) *evm.UnsignedImportTx {
	importedAmount := make(map[ids.ID]uint64)
	outs := make([]evm.EVMOutput, 0, len(importedAmount))
	for assetID, amount := range importedAmount {
		outs = append(outs, evm.EVMOutput{
			Address: args.To,
			Amount:  amount,
			AssetID: assetID,
		})
	}

	utx := &evm.UnsignedImportTx{
		NetworkID:      args.NetworkID,
		BlockchainID:   args.BlockchainID,
		SourceChain:    args.SourceChainID,
		ImportedInputs: myAvax.TransferableInputsrFromUTXOs(args.AtomicUTXOs), // Inputs that consume UTXOs produced on the chain
		Outs:           outs,                                                  // Outputs
	}
	return utx
}
