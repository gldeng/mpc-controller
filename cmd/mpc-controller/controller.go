package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/contract/reconnector"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/support/keygen"
	"github.com/avalido/mpc-controller/tasks/joining"
	"github.com/avalido/mpc-controller/tasks/staking"
	"github.com/avalido/mpc-controller/utils/bytes"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"golang.org/x/sync/errgroup"
	"math/big"
	"reflect"
	"time"

	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/queue"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/avalido/mpc-controller/storage/badgerDB"
	"github.com/avalido/mpc-controller/support/participant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
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
		name := reflect.TypeOf(service).String()

		service := service
		g.Go(func() error {
			fmt.Printf("Service %q started at: %v \n", name, time.Now())
			if err := service.Start(ctx); err != nil {
				return errors.Wrapf(err, "failed starting service: %v", name)
			}
			fmt.Printf("Service %q stopped at: %v\n", name, time.Now())
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errors.WithStack(err)
	}

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

	// Create mpcClient
	mpcClient, _ := core.NewMpcClient(myLogger, config.MpcServerUrl)

	// Parse private myPrivKey
	myPrivKey, err := crypto.HexToECDSA(config.ControllerKey)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to parse controller private key %q", config.ControllerKey))
	}

	// Parse public myPrivKey
	myPubKeyBytes := myCrypto.MarshalPubkey(&myPrivKey.PublicKey)[1:]
	myPubKeyHex := bytes.BytesToHex(myPubKeyBytes)
	myPubKeyHash := hash256.FromBytes(myPubKeyBytes)

	// Convert chain ID
	chainId := big.NewInt(config.ChainId)

	// Create controller transaction signer
	signer, err := bind.NewKeyedTransactorWithChainID(myPrivKey, chainId)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to create controller transaction signer", logger.Field{"error", err}))
	}

	// Create eth rpc client
	ethRpcClient, err := ethclient.Dial(config.EthRpcUrl)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to connect eth rpc client, url: %q", config.EthRpcUrl))
	}

	// Create eth ws client
	ethWsClient, err := ethclient.Dial(config.EthWsUrl)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to connect eth ws client, url: %q", config.EthWsUrl))
	}

	// Create P-Chain issue client
	pChainIssueCli := platformvm.NewClient(config.PChainIssueUrl)

	// Create contract filterer reconnector
	filterReconnector := reconnector.ContractFilterReconnector{
		Logger:    myLogger,
		Updater:   &network.EthClientDialerImpl{myLogger, config.EthWsUrl, ethWsClient},
		Publisher: myDispatcher,
	}

	cacheWrapper := cache.CacheWrapper{
		Dispatcher: myDispatcher,
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

	keygenMaster := keygen.KeygenMaster{
		Logger:            myLogger,
		ContractAddr:      contractAddr,
		Dispatcher:        myDispatcher,
		KeygenDoner:       mpcClient,
		Storer:            myDB,
		PChainIssueClient: pChainIssueCli,
		MyPubKeyHashHex:   myPubKeyHash.Hex(),
		Transactor:        ethRpcClient,
		Signer:            signer,
		Receipter:         ethRpcClient,
	}

	joiningMaster := joining.JoiningMaster{
		Logger:          myLogger,
		ContractAddr:    contractAddr,
		MyPubKeyHashHex: myPubKeyHash.Hex(),
		MyIndexGetter:   &cacheWrapper,
		Dispatcher:      myDispatcher,
		Signer:          signer,
		Receipter:       ethRpcClient,
		Transactor:      ethRpcClient,
	}

	stakingMaster := staking.StakingMaster{
		Logger:          myLogger,
		ContractAddr:    contractAddr,
		MyPubKeyHashHex: myPubKeyHash.Hex(),
		Dispatcher:      myDispatcher,
		NetworkContext:  networkCtx(config),
		Cache:           &cacheWrapper,
		SignDoner:       mpcClient,
		Verifyier:       nil,
		Noncer:          ethRpcClient,
	}

	controller := MpcController{
		ID: config.ControllerId,
		Services: []Service{
			&cacheWrapper,
			&filterReconnector,
			&participantMaster,
			&keygenMaster,
			&joiningMaster,
			&stakingMaster,
		},
	}

	return &controller
}

func networkCtx(config *config.Config) chain.NetworkContext {
	// Convert C-Chain ID
	cchainID, err := ids.FromString(config.CChainId)
	if err != nil {
		panic(errors.Wrap(err, "Failed to convert C-Chain ID"))
	}

	// Convert chain ID
	chainIdBigInt := big.NewInt(config.ChainId)

	// Convert AVAX assetId ID
	assetId, err := ids.FromString(config.AvaxId)
	if err != nil {
		panic(errors.Wrap(err, "Failed to convert AVAX assetId"))
	}

	// Create NetworkContext
	networkCtx := chain.NewNetworkContext(
		config.NetworkId,
		cchainID,
		chainIdBigInt,
		avax.Asset{
			ID: assetId,
		},
		config.ImportFee,
		config.GasPerByte,
		config.GasPerSig,
		config.GasFixed,
	)
	return networkCtx
}
