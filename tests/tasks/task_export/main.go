package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/tasks/c2p"
	"github.com/avalido/mpc-controller/utils/backoff"
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

	mpcClient, err := core.NewSimulatingMpcClient("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")

	panicIfError(err)
	config := pool.TaskContextImpConfig{
		Logger:     logger.Default(),
		Host:       "34.172.25.188",
		Port:       9650,
		SslEnabled: false,
		NetworkContext: chain.NewNetworkContext(
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
		MpcClient: mpcClient,
	}
	ctx, err := pool.NewTaskContextImp(config)
	panicIfError(err)
	quorum := c2p.QuorumInfo{
		ParticipantPubKeys: nil,
		PubKey:             mpcClient.UncompressedPublicKeyBytes(),
	}
	task, err := c2p.NewExportFromCChain("abc", *big.NewInt(100), quorum)
	panicIfError(err)
	nextTasks, err := task.Next(ctx)
	panicIfError(err)
	nextTasks, err = task.Next(ctx)
	panicIfError(err)
	time.Sleep(5 * time.Second)
	fmt.Printf("TxID is %v\n", task.TxID.String())
	backoff.RetryFnExponential10Times(logger.Default(), context.Background(), 1*time.Second, 120*time.Second, func() (retry bool, err error) {
		nextTasks, err = task.Next(ctx)
		if err != nil {
			return false, err
		}
		if len(nextTasks) > 0 {
			return true, nil
		}
		fmt.Printf("Task IsDone: %v\n", task.IsDone())
		return false, nil
	})

	fmt.Printf("next is %v\n", nextTasks)
}
