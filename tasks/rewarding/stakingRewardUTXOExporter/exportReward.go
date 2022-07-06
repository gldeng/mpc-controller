package stakingRewardUTXOExporter

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/signer"
	"github.com/avalido/mpc-controller/utils/crypto/secp256k1r"
	portIssuer "github.com/avalido/mpc-controller/utils/port/issuer"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/avalido/mpc-controller/utils/port/txs/cchain"
	"github.com/avalido/mpc-controller/utils/port/txs/pchain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type Args struct {
	NetworkID   uint32
	PChainID    ids.ID
	CChainID    ids.ID
	PChainAddr  ids.ShortID
	CChainArr   common.Address
	RewardUTXOs []*avax.UTXO

	SignDoner   core.SignDoner
	SignReqArgs *signer.SignRequestArgs

	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.CChainIssuer
}

func exportReward(ctx context.Context, args *Args) ([2]ids.ID, error) {
	amountToExport := args.RewardUTXOs[0].Out.(*secp256k1fx.TransferOutput).Amount()
	myExportTxArgs := &pchain.Args{
		NetworkID:          args.NetworkID,
		BlockchainID:       args.PChainID,
		DestinationChainID: args.CChainID,
		Amount:             amountToExport, // todo: to be tuned
		To:                 args.PChainAddr,
		UTXOs:              args.RewardUTXOs,
	}

	myImportTxArgs := &cchain.Args{
		NetworkID:     args.NetworkID,
		BlockchainID:  args.CChainID,
		SourceChainID: args.PChainID,
		To:            args.CChainArr,
		//BaseFee:      *big.Int todo: tobe consider
	}

	myTxs := &Txs{
		UnsignedExportTxArgs: myExportTxArgs,
		UnsignedImportTx:     myImportTxArgs,
	}

	mySigner := &signer.Signer{args.SignDoner, *args.SignReqArgs}
	myVerifier := &secp256k1r.Verifier{PChainAddress: args.PChainAddr}
	myIssuer := &portIssuer.Issuer{args.CChainIssueClient, args.PChainIssueClient, portIssuer.P2C}
	myPorter := porter.Porter{myTxs, mySigner, myIssuer, myVerifier}

	txIds, err := myPorter.SignAndIssueTxs(ctx)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	return txIds, nil
}
