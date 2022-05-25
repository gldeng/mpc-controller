package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/task"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	configFile = "configFile"
)

func mpcController(c *cli.Context) error {
	configImpl := config.ParseConfigFromFile(c.String(configFile))
	configInterface := config.InitConfig(configImpl)

	logger.DevMode = configInterface.IsDevMode()
	log := logger.Default()

	staker := task.NewStaker(log, configInterface.CChainIssueClient(), configInterface.PChainIssueClient())
	storer := storage.New(log, configImpl.DatabasePath())

	m, err := task.NewTaskManager(log, configInterface, storer, staker)
	if err != nil {
		return errors.Wrap(err, "Failed to create task-manager for mpc-controller")
	}

	// Start the mpc-controller.
	ctx, shutdown := context.WithCancel(context.Background())
	go func() {
		err = m.Start(ctx)
		if err != nil {
			fmt.Printf("Failed to run mpc-controller, error: %+v", err)
			os.Exit(1)
		}
	}()

	// Handle graceful shutdown.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdown()
	time.Sleep(time.Second * 10) // wait for while

	return nil
}

// todo: keystore to strength private key security
// todo: automatic panic recover
// todo: distributed trace, log and monitor
// todo: deal with error: invalid nonce
// todo: check and sync participant upon startup, there ere maybe groups created during mpc-controller downtime.
// todo: add mpc-controller version info
// todo: mechanism to check result from mpc-server and resume task on mpc-controller startup
// todo: history even track for mpc-coordinator smart contract.

func main() {
	app := &cli.App{
		Name:  "mpc-controller",
		Usage: "Handles the MPC operations needed for Avalanche",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     configFile,
				Required: true,
				Usage:    "The config file path for mpc-controller",
			},
		},
		Action: mpcController,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Failed to run mpc-controller, error: %+v", err)
	}
}
