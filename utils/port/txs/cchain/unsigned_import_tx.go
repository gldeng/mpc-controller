package cchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Args struct {
	NetworkID   uint32
	ChainID     ids.ID         // chain to import from
	To          common.Address // Address of recipient
	BaseFee     *big.Int       // fee to use post-AP3
	AtomicUTXOs []*avax.UTXO   // UTXOs to spend
}

func UnsignedImportTx(args *Args) *evm.UnsignedImportTx {
	importedInputs := []*avax.TransferableInput{}

	importedAmount := make(map[ids.ID]uint64)
	for _, utxo := range args.AtomicUTXOs {
		importedInputs = append(importedInputs, &avax.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  utxo.Asset,
			In:     utxo.Out.(*secp256k1fx.TransferInput), // todo: fix it
		})
	}

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
		BlockchainID:   args.ChainID,
		Outs:           outs,
		ImportedInputs: importedInputs,
		SourceChain:    args.ChainID,
	}
	return utx
}
