package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mpcclient"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/taskcontext"
	"github.com/avalido/mpc-controller/tasks/stake"
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
	logger.DevMode = true
	logger.UseConsoleEncoder = true

	mpcClient, err := mpcclient.NewSimulatingMpcClient("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	panicIfError(err)

	config := core.Config{
		Host:              "34.172.25.188",
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
	services := core.NewServicePack(config, logger.Default(), mpcClient, db)
	ctx, err := taskcontext.NewTaskContextImp(services)

	quorum := types.QuorumInfo{
		ParticipantPubKeys: nil,
		PubKey:             mpcClient.UncompressedPublicKeyBytes(), // NOTE: use the compressed ones for true MPC signing request
	}

	timestamp := uint64(time.Now().Unix()) + 5*60
	task, err := stake.NewInitialStake(&stake.Request{
		ReqNo:     0,
		TxHash:    common.Hash{},
		PubKey:    mpcClient.UncompressedPublicKeyBytes(),
		NodeID:    "NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5",
		Amount:    "100000000000000000000",
		StartTime: timestamp,
		EndTime:   timestamp + 2*60*60,
	}, quorum)

	panicIfError(err)
	nextTasks, err := task.Next(ctx)
	panicIfError(err)
	nextTasks, err = task.Next(ctx)
	panicIfError(err)
	time.Sleep(5 * time.Second)
	backoff.RetryFnConstant(logger.Default(), context.Background(), 1000, 1*time.Second, func() (retry bool, err error) {
		nextTasks, err = task.Next(ctx)
		if err != nil {
			return false, err
		}
		if len(nextTasks) > 0 {
			return true, nil
		}
		fmt.Printf("AddDelegator IsDone: %v\n", task.IsDone())
		return !task.IsDone(), nil
	})
	fmt.Printf("next is %v\n", nextTasks)
}
