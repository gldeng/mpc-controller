package task

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"time"
)

const (
	ADDR_CCHAIN = "0x8db97c7cece249c2b98bdc0226cc4c2a57bf52fc"
)

//
var cChainAddress = common.HexToAddress(ADDR_CCHAIN)

// todo: refactor with config provided

func doStake(task *StakeTask) error {
	logger.DevMode = true

	client, err := ethclient.Dial("http://localhost:9650/ext/bc/C/rpc")

	nonce, err := client.NonceAt(context.Background(), cChainAddress, nil)
	logger.Debug("get nonce at account", logger.Field{"Address", cChainAddress}, logger.Field{"nonce", nonce})

	tx1, err := task.GetSignedExportTx()
	if err != nil {
		return errors.WithStack(err)
	}

	cclient := evm.NewClient("http://localhost:9650", "C")
	txId1, err := cclient.IssueTx(context.Background(), tx1.Bytes())

	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("ExportTx %v\n", txId1)
	time.Sleep(time.Second * 2)
	pclient := platformvm.NewClient("http://localhost:9650")
	tx2, err := task.GetSignedImportTx()
	if err != nil {
		return errors.WithStack(err)
	}
	txId2, err := pclient.IssueTx(context.Background(), tx2.Bytes())
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("ImportTx %v\n", txId2)
	time.Sleep(time.Second * 2)
	tx3, err := task.GetSignedAddDelegatorTx()
	if err != nil {
		return errors.WithStack(err)
	}

	txId3, err := pclient.IssueTx(context.Background(), tx3.Bytes())
	if err != nil {
		logger.Error("Task-manager Failed to issue Tx to P-Chain",
			logger.Field{"error", err})
		return errors.WithStack(err)
	}
	fmt.Printf("----------------AddDelegatorTx %v\n", txId3)

	return nil
}
