// use UnsignedImportTx.GasUsed instead

package cchain

import (
	"context"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/avalido/mpc-controller/logger"
	"github.com/pkg/errors"
	"math/big"
)

type EstimateAtomicTxFeeArgs struct {
	Logger logger.Logger
	Client ethclient.Client
	AtomicTxFeeArgs
}

type AtomicTxFeeArgs struct {
	UnsignedTxBytes []byte
	NumSignatures   uint64

	PerUnsignedTxByteFee uint64
	PerSignatureFee      uint64
	PerAtomicTxFee       uint64
}

func EstimateAtomicTxFee(ctx context.Context, args *EstimateAtomicTxFeeArgs) (uint64, error) {
	baseFeeBig, err := args.Client.EstimateBaseFee(ctx)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	AtomicTxBaseFeeBig := new(big.Int).Div(baseFeeBig, big.NewInt(1_000_000_000))

	return AtomicTxFee(AtomicTxBaseFeeBig.Uint64(), &args.AtomicTxFeeArgs), nil

}

func AtomicTxFee(baseFee uint64, args *AtomicTxFeeArgs) uint64 {
	return AtomicTxGasFee(args) * baseFee
}

func AtomicTxGasFee(args *AtomicTxFeeArgs) uint64 {
	unsignedTxLen := len(args.UnsignedTxBytes)
	return args.PerUnsignedTxByteFee*uint64(unsignedTxLen) + args.PerSignatureFee*args.NumSignatures + args.PerAtomicTxFee
}
