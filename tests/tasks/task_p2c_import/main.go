package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mpcclient"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/taskcontext"
	"github.com/avalido/mpc-controller/tasks/p2c"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func idFromString(str string) ids.ID {
	id, err := ids.FromString(str)
	panicIfError(err)
	return id
}

func main() {

	// CHANGE ME -->
	host := "172.21.128.1"
	hexTx := "0x0000000000120000053900000000000000000000000000000000000000000000000000000000000000000000000000000001f3ad10ac70eb938deb2044be72746e4de581ad87eb69d6a5a2286e2da2a5aaa80000000017cc8b1578ba383544d163958822d8abd3849bb9dfabe39fcbc3e7ee8811fe2f00000005000000003b9aca00000000010000000000000000d41e6504141f003a4f1689073f198dbc128e5d4692f72b7e48a29728a42653b50000000117cc8b1578ba383544d163958822d8abd3849bb9dfabe39fcbc3e7ee8811fe2f00000007000000003b8b87c0000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c0000000100000009000000012b5606a457d7e53438a5b008f38a4f01174024bc064d2d5c599e70ad11e0acb06a510821e8c11376a5a40511a3e61d2612c0b3ace3f4498da405db19123184560171347840"
	toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	// <--

	mpcClient, err := mpcclient.NewSimulatingClient("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")

	panicIfError(err)
	config := core.Config{
		Host:              host,
		Port:              9650,
		SslEnabled:        false,
		MpcManagerAddress: common.Address{},
		NetworkContext: core.NewNetworkContext(
			1337,
			idFromString("2cRHidGTGMgWSMQXVuyqB86onp69HTtw6qHsoHvMjk9QbvnijH"),
			big.NewInt(43112),
			avax.Asset{
				ID: idFromString("BUuypiq2wyuLMvyhzFXcPyxPMCgSp7eeDohhQRqTChoBjKziC"),
			},
			1000000,
			1000000,
			1,
			1000,
			10000,
			300,
		),
		MyPublicKey: common.Hex2Bytes("3217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8bdd1c124d85f6c0b10b6ef24460ff4acd0fc2cd84bd5b9c7534118f472d0c7a1"),
	}

	db := storage.NewInMemoryDb()
	txIndex := core.NewInMemoryTxIndex()
	services := core.NewServicePack(config, logger.Default(), mpcClient, db, txIndex)
	ctx, err := taskcontext.NewTaskContextImp(services)
	panicIfError(err)
	quorum := types.QuorumInfo{
		ParticipantPubKeys: nil,
		PubKey:             mpcClient.UncompressedPublicKeyBytes(),
	}

	bytesTx, err := formatting.Decode(formatting.Hex, hexTx)
	panicIfError(err)

	tx, err := txs.Parse(txs.Codec, bytesTx)
	panicIfError(err)

	task, err := p2c.NewImportIntoCChain(core.FlowId{
		Tag:         "abc",
		RequestHash: types.RequestHash{},
	}, quorum, tx, toAddress)
	panicIfError(err)
	nextTasks, err := task.Next(ctx)
	panicIfError(err)
	nextTasks, err = task.Next(ctx)
	panicIfError(err)
	time.Sleep(5 * time.Second)
	backoff.RetryFnExponential10Times(logger.Default(), context.Background(), 1*time.Second, 120*time.Second, func() (retry bool, err error) {
		nextTasks, err = task.Next(ctx)
		if err != nil {
			panic(err)
		}
		if !task.IsDone() && !task.FailedPermanently() {
			return true, nil
		}
		fmt.Printf("Task IsDone: %v\n", task.IsDone())
		return false, nil
	})

	fmt.Printf("TxID is %v\n", task.TxID.String())
	fmt.Printf("next is %v\n", nextTasks)
}
