package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	addDelegator "github.com/avalido/mpc-controller/tasks/adddelegator"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/testingutils"
	"github.com/pkg/errors"
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

func fakeStakeParam() (*addDelegator.StakeParam, error) {
	nodeID, _ := ids.NodeIDFromString("NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5")
	utxos, err := testingutils.GetRewardUTXOs("http://34.172.25.188:9650", "cxbA4wytAUWTRmNyqfYQHnHdR8vYthyeCrDFWEQULiUHPyVu2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get utxos")
	}

	if len(utxos) == 0 {
		return nil, errors.New("no utxo found")
	}
	param := addDelegator.StakeParam{
		NodeID:    nodeID,
		StartTime: 1663315662,
		EndTime:   1694830062,
		UTXOs:     utxos,
	}
	return &param, nil
}

func main() {
	logger.DevMode = true
	logger.UseConsoleEncoder = true

	mpcClient, err := core.NewSimulatingMpcClient("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	panicIfError(err)

	ctx := &core.TaskContextImp{
		Logger: logger.Default(),
		Network: chain.NewNetworkContext(
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
		MpcClient:    mpcClient,
		PChainClient: nil, // TODO:
	}

	quorum := types.QuorumInfo{
		ParticipantPubKeys: nil,
		PubKey:             mpcClient.UncompressedPublicKeyBytes(),
	}

	stakeParam, err := fakeStakeParam()
	panicIfError(err)

	task, err := addDelegator.NewAddDelegator("0xc02b59f772cb23a75b6ffb9f7602ba25fdd5d8e75ad88efcc013fec2c63b0895", quorum, stakeParam)
	panicIfError(err)
	nextTasks, err := task.Next(ctx)
	panicIfError(err)
	nextTasks, err = task.Next(ctx)
	panicIfError(err)
	time.Sleep(5 * time.Second)
	backoff.RetryFnExponential10Times(logger.Default(), context.Background(), 1*time.Second, 120*time.Second, func() (retry bool, err error) {
		nextTasks, err = task.Next(ctx)
		if err != nil {
			return false, err
		}
		if len(nextTasks) > 0 {
			return true, nil
		}
		fmt.Printf("ExportTask IsDone: %v\n", task.IsDone())
		return false, nil
	})
	fmt.Printf("next is %v\n", nextTasks)
}
