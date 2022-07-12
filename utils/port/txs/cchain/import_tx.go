package cchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/coreth/plugin/evm"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/ethereum/go-ethereum/common"
)

type ImportTxArgs struct {
	NetworkID     uint32         // ID of the network on which this tx was issued
	BlockchainID  ids.ID         // ID of this blockchain.
	SourceChainID ids.ID         // Which chain to consume the funds from
	To            common.Address // Address of recipient
	//BaseFee      *big.Int       // fee to use post-AP3 // todo: consider this kind of fee
	AtomicUTXOs []*avax.UTXO // UTXOs to spend
}

func ImportTx(args *ImportTxArgs) *evm.UnsignedImportTx {
	mpcUTXOs := myAvax.MpcUTXOsFromUTXOs(args.AtomicUTXOs)
	importedAmounts := make(map[ids.ID]uint64)
	for _, mpcUTXO := range mpcUTXOs {
		importedAmounts[mpcUTXO.Asset] = mpcUTXO.Out.Amt
	}

	outs := make([]evm.EVMOutput, 0, len(importedAmounts))
	for assetID, amount := range importedAmounts {
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
