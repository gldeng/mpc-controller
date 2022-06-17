package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/utils/bytes"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"

	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/queue"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/avalido/mpc-controller/storage/badgerDB"
	"github.com/avalido/mpc-controller/support/participant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Start(ctx context.Context) error
}

type MpcController struct {
	ID       string
	Services []Service
}

func (c *MpcController) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, service := range c.Services {
		service := service
		g.Go(func() error {
			return service.Start(ctx)
		})
	}

	fmt.Printf("%v services started.\n", c.ID)
	if err := g.Wait(); err != nil {
		return errors.WithStack(err)
	}

	fmt.Printf("%v services closed.\n", c.ID)
	return nil
}

func NewController(ctx context.Context, c *cli.Context) *MpcController {
	// Parse config
	config := config.ParseFile(c.String(configFile))

	// Create logger
	logger.DevMode = config.EnableDevMode
	myLogger := logger.Default()

	// Create database
	myDB := &badgerDB.BadgerDB{
		Logger: myLogger,
		DB:     badgerDB.NewBadgerDBWithDefaultOptions(config.BadgerDbPath),
	}

	// Create dispatcher
	myDispatcher := dispatcher.NewDispatcher(ctx, myLogger, queue.NewArrayQueue(1024), 1024)

	// Get MpcManager contract address
	contractAddr := common.HexToAddress(config.MpcManagerAddress)

	//// Create mpcClient
	//mpcClient, _ := core.NewMpcClient(myLogger, config.MpcServerUrl)

	// Parse private myPrivKey
	myPrivKey, err := crypto.HexToECDSA(config.ControllerKey)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to parse private myPrivKey", logger.Field{"error", err}))
	}

	// Parse public myPrivKey
	myPubKeyBytes := myCrypto.MarshalPubkey(&myPrivKey.PublicKey)[1:]
	myPubKeyHex := bytes.BytesToHex(myPubKeyBytes)
	myPubKeyHash := hash256.FromBytes(myPubKeyBytes)

	//// Convert chain ID
	//chainId := big.NewInt(config.ChainId)

	//// Create controller transaction signer
	//signer, err := bind.NewKeyedTransactorWithChainID(myPrivKey, chainId)
	//if err != nil {
	//	panic(errors.Wrapf(err, "Failed to create controller transaction signer", logger.Field{"error", err}))
	//}

	// Create eth rpc client
	ethRpcClient, err := ethclient.Dial(config.EthRpcUrl)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to connect eth rpc client", logger.Field{"error", err}))
	}

	participantMaster := participant.ParticipantMaster{
		Logger:          myLogger,
		MyPubKeyHex:     myPubKeyHex,
		MyPubKeyHashHex: myPubKeyHash.Hex(),
		MyPubKeyBytes:   myPubKeyBytes,
		ContractAddr:    contractAddr,
		ContractCaller:  ethRpcClient,
		Dispatcher:      myDispatcher,
		Storer:          myDB,
	}

	controller := MpcController{
		ID: config.ControllerId,
		Services: []Service{
			&participantMaster,
		},
	}

	return &controller
}
