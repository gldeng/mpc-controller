package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/router"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/subscriber"
	"github.com/avalido/mpc-controller/tasks/ethlog"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
)

const (
	fnHost              = "host"
	fnPort              = "port"
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
		EthWsURL:          fmt.Sprintf("ws://%s:%v/ext/bc/C/ws", c.String(fnHost), c.Int(fnPort)),
		MpcManagerAddress: common.HexToAddress(c.String(fnMpcManagerAddress)),
	}, q)

	coreConfig := core.Config{
		Host:           c.String(fnHost),
		Port:           int16(c.Int(fnPort)),
		SslEnabled:     false, // TODO: Add argument
		NetworkContext: chain.NetworkContext{},
	}

	db := &storage.InMemoryDb{}

	services := core.NewServicePack(coreConfig, myLogger, nil, db)

	ehContext, err := core.NewEventHandlerContextImp(services)
	if err != nil {
		return err
	}

	rt, _ := router.NewRouter(q, ehContext)
	rt.AddHandler(printLog)

	rc := &ethlog.RequestCreator{}
	rt.AddLogEventHandler(rc)
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
				Name:        fnHost,
				Required:    true,
				Usage:       "The host of the avalanche rpc service.",
				DefaultText: "localhost",
			},

			&cli.IntFlag{
				Name:        fnPort,
				Required:    true,
				Usage:       "The port of the avalanche rpc service.",
				DefaultText: "9650",
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
