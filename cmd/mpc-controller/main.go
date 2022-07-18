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
