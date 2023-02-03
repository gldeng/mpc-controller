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
	utxoHex := "0x0000f017f8b4f679b54cb490d4bb42bdd00f960629b74b2574d9cfbbf176bc7efc210000000017cc8b1578ba383544d163958822d8abd3849bb9dfabe39fcbc3e7ee8811fe2f00000007000000003b9aca00000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29ca04c263f"
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
	task, err := p2c.NewP2C(core.FlowId{
		Tag:         "abc",
		RequestHash: types.RequestHash{},
	}, quorum, *parseUtxo(utxoHex), toAddress)
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
		if !task.IsDone() && !task.FailedPermanently() {
			return true, nil
		}
		fmt.Printf("ExportTask IsDone: %v\n", task.ExportTask.IsDone())
		return false, nil
	})
	fmt.Printf("Export TxID is %v\n", task.ExportTask.TxID.String())
	backoff.RetryFnExponential10Times(logger.Default(), context.Background(), 1*time.Second, 120*time.Second, func() (retry bool, err error) {
		nextTasks, err = task.Next(ctx)
		if err != nil {
			return false, err
		}
		if !task.IsDone() && !task.FailedPermanently() {
			return true, nil
		}
		fmt.Printf("ImportTask IsDone: %v\n", task.ImportTask.IsDone())
		return false, nil
	})
	fmt.Printf("Import TxID is %v\n", task.ImportTask.TxID.String())
	fmt.Printf("next is %v\n", nextTasks)
}

func parseUtxo(hexRep string) *avax.UTXO {
	bUtxo, err := formatting.Decode(formatting.Hex, hexRep)
	panicIfError(err)

	utxo := &avax.UTXO{}
	_, err = txs.Codec.Unmarshal(bUtxo, utxo)
	panicIfError(err)
	return utxo
}
