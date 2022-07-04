package cchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Args struct {
	NetworkID   uint32
	ChainID     ids.ID         // chain to import from
	To          common.Address // Address of recipient
	BaseFee     *big.Int       // fee to use post-AP3 // todo: consider this kind of fee
	AtomicUTXOs []*myAvax.UTXO // UTXOs to spend
}

func UnsignedImportTx(args *Args) *evm.UnsignedImportTx {
	importedInputs := []*avax.TransferableInput{}

	importedAmount := make(map[ids.ID]uint64)
	for _, utxo := range args.AtomicUTXOs {
		importedInputs = append(importedInputs, &avax.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  utxo.Asset,
			In: &secp256k1fx.TransferInput{
				Amt: utxo.Out.Amt, // todo: to adjust it to import fee
				Input: secp256k1fx.Input{
					SigIndices: []uint32{0}, // todo: to adjust?
				},
			},
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
