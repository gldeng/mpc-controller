package main

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
	password   = "password"
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
			&cli.StringFlag{
				Name:     password,
				Required: true,
				Usage:    "The password to decrypt mpc-controller key",
			},
		},
		Action: RunMpcController,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Failed to run mpc-controller, error: %+v", err)
	}
}

//## Todo List
//- enhanced keystore to strength private key security
//- automatic panic recover
//- distributed trace, log and monitor
//- deal with casual error: invalid nonce and nonce misused
//- check and sync participant upon startup, there ere maybe groups created during mpc-controller downtime.
//- add mpc-controller version info
//- mechanism to check result from mpc-server and resume task on mpc-controller startup
//- history even track for mpc-coordinator smart contract.
//- log rotation with lumberjack: https://github.com/natefinch/lumberjack
//- add main_test.go
//- apply confluentinc/bincover: https://github.com/confluentinc/bincover
//- restore data on startup
//- automate tracking balance of addresses that receive principal and reward.
//- take measures to deal with failed tasks
//- take measures to avoid double-spend, maybe introduce SPE(single-participant-execution) strategy or consensus
//- take measures to deal with package lost and disorder arrival
//- ...
