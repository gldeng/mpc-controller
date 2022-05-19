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

	fmt.Printf("---------- Starting mpc-controller %s\n", configImpl.ControllerId())

	// Start the mpc-controller.
	go func() {
		err = m.Start()
		if err != nil {
			fmt.Println("Failed to start mpc-manager!")
			panic(err)
		}
	}()

	// Handle graceful shutdown.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("---------- Gracefully shutting down the mpc-controller...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = m.Shutdown(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to shutdown mpc-controller")
	}

	fmt.Println("---------- The mpc-controller shut down.")

	return nil
}

// todo: keystore to strength private key security
// todo: listening signals
// todo: elegant shutdown
// todo: automatic panic recover
// todo: distributed trace, log and monitor
// todo: deal with gorutine leak
// todo: deal with error: invalid nonce
// todo: check and sync participant upon startup, there ere maybe groups created during mpc-controller downtime.
// todo: add mpc-controller version info

func main() {
	logger.DevMode = true // remove this line later
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
	logger.FatalOnError(err, "Failed to run mpc-controller.", logger.Field{"error", err})
}
