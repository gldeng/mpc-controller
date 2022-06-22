package main

// todo: keystore to strength private key security
// todo: automatic panic recover
// todo: distributed trace, log and monitor
// todo: deal with error: invalid nonce
// todo: check and sync participant upon startup, there ere maybe groups created during mpc-controller downtime.
// todo: add mpc-controller version info
// todo: mechanism to check result from mpc-server and resume task on mpc-controller startup
// todo: history even track for mpc-coordinator smart contract.
// todo: log rotation with lumberjack: https://github.com/natefinch/lumberjack

// todo: add main_test.go
// todo: apply confluentinc/bincover: https://github.com/confluentinc/bincover
// todo: restore data on startup

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	configFile = "configFile"
)

func RunMpcController(c *cli.Context) error {
	shutdownCtx, shutdown := context.WithCancel(context.Background())

	controller := NewController(shutdownCtx, c)

	// Handle graceful shutdown.
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		shutdown()
	}()

	if err := controller.Run(shutdownCtx); err != nil {
		return errors.WithStack(err)
	}

	time.Sleep(time.Second * 5)
	return nil
}

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
		Action: RunMpcController,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Failed to run mpc-controller, error: %+v", err)
	}
}
