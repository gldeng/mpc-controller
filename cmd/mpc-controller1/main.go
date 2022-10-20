package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/router"
	"github.com/avalido/mpc-controller/subscriber"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
)

const (
	fnRpcUrl            = "rpc-url"
	fnMpcManagerAddress = "mpc-manager-address"
)

func printLog(event interface{}) {
	evt, ok := event.(types.Log)
	if !ok {
		return
	}
	fmt.Printf("Received event log %v\n", evt)
}

func runController(c *cli.Context) error {

	myLogger := logger.Default()
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	q := goconcurrentqueue.NewFIFO()

	sub, err := subscriber.NewSubscriber(shutdownCtx, myLogger, &subscriber.Config{
		EthWsURL:          c.String(fnRpcUrl),                                 // "ws://34.172.25.188:9650/ext/bc/C/ws",
		MpcManagerAddress: common.HexToAddress(c.String(fnMpcManagerAddress)), // 0x354f6dA5Bfca855021b6AbE4f138AD94bEB688D2
	}, q)

	rt, _ := router.NewRouter(q)
	rt.AddHandler(printLog)
	err = sub.Start()
	if err != nil {
		return err
	}
	err = rt.Start()
	if err != nil {
		return err
	}

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		shutdown()
		rt.Close()
		sub.Close()
	}()

	<-shutdownCtx.Done()

	return nil
}

func main() {

	app := &cli.App{
		Name:  "mpc-controller",
		Usage: "Handles the MPC operations needed for Avalanche",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     fnRpcUrl,
				Required: true,
				Usage:    "The url of the rpc node.",
			},
			&cli.StringFlag{
				Name:     fnMpcManagerAddress,
				Required: true,
				Usage:    "The address of the deployed MpcManager contract.",
			},
		},
		Action: runController,
	}

	fmt.Printf("Starting process: %v\n", os.Getpid())

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Failed to run controolerr, error: %+v", err)
	}
}
