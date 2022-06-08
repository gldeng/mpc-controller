package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/contract/wrappers"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/services/group"
	"github.com/avalido/mpc-controller/services/keygen"
	"github.com/avalido/mpc-controller/services/stake"
	"github.com/avalido/mpc-controller/storage"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type MpcController struct {
	ID       string
	Services []mpc_controller.MpcControllerService
}

func (c *MpcController) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, service := range c.Services {
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

func NewController(c *cli.Context) *MpcController {
	// Initiate config
	myLogger, myConfig, myStorer := initConfig(c)

	// Build services
	privateKey := myConfig.ControllerKey()
	pubKeyBytes := myCrypto.MarshalPubkey(&privateKey.PublicKey)[1:]
	pubKeyHex := common.Bytes2Hex(pubKeyBytes)
	pubKeyHash := crypto.Keccak256Hash(pubKeyBytes)
	pubKeyHashHex := pubKeyHash.Hex()

	signer := myConfig.ControllerSigner()

	mpcManagerCallerWrapper := wrappers.MpcManagerCallerWrapper{myLogger, &myConfig.CoordinatorBoundInstance().MpcManagerCaller}
	mpcManagerFilterWrapper := wrappers.MpcManagerFilterWrapper{myLogger, &myConfig.CoordinatorBoundListener().MpcManagerFilterer}
	mpcManagerTransactorWrapper := wrappers.MpcManagerTransactorWrapper{myLogger, &myConfig.CoordinatorBoundInstance().MpcManagerTransactor}

	mpcClient := myConfig.MpcClient()

	rpcEthClient := myConfig.EthRpcClient()
	rpcEthClientWrapper := chain.RpcEthClientWrapper{myLogger, rpcEthClient}

	cChainIssueClient := myConfig.CChainIssueClient()
	pChainIssueClient := myConfig.PChainIssueClient()

	// Build group service
	groupService := group.Group{
		PubKeyStr:               pubKeyHex,
		PubKeyBytes:             pubKeyBytes,
		PubKeyHashStr:           pubKeyHashHex,
		Logger:                  myLogger,
		CallerGetGroup:          &mpcManagerCallerWrapper,
		WatcherParticipantAdded: &mpcManagerFilterWrapper,

		StorerStoreGroupInfo:       myStorer,
		StorerStoreParticipantInfo: myStorer,
	}

	// Build keygen service
	keygenService := keygen.Keygen{
		PubKeyHashHex:                pubKeyHashHex,
		Logger:                       myLogger,
		MpcClientKeygen:              mpcClient,
		MpcClientResult:              mpcClient,
		WatcherKeygenRequestAdded:    &mpcManagerFilterWrapper,
		TransactorReportGeneratedKey: &mpcManagerTransactorWrapper,

		StorerGetGroupIds:         myStorer,
		StorerLoadParticipantInfo: myStorer,

		StorerLoadKeygenRequestInfo:    myStorer,
		StorerStoreGeneratedPubKeyInfo: myStorer,

		StorerLoadGroupInfo:          myStorer,
		StorerStoreKeygenRequestInfo: myStorer,

		EthClientTransactionReceipt: rpcEthClient,

		Signer: signer,
	}

	// Build stake service

	stakeService := stake.Manager{
		Logger: myLogger,

		Staker: &stake.Staker{myLogger, cChainIssueClient, pChainIssueClient},

		MpcClientSign:   mpcClient,
		MpcClientResult: mpcClient,
		NetworkContext:  *myConfig.NetworkContext(),

		TransactorJoinRequest:      &mpcManagerTransactorWrapper,
		WatcherStakeRequestAdded:   &mpcManagerFilterWrapper,
		WatcherStakeRequestStarted: &mpcManagerFilterWrapper,

		EthClientTransactionReceipt: &rpcEthClientWrapper,
		EthClientNonceAt:            &rpcEthClientWrapper,
		EthClientBalanceAt:          &rpcEthClientWrapper,

		StorerGetParticipantIndex:     myStorer,
		StorerGetPariticipantKeys:     myStorer,
		StorerGetPubKeys:              myStorer,
		StorerLoadGeneratedPubKeyInfo: myStorer,

		Signer:     signer,
		PubKeyHash: pubKeyHash,
	}

	controller := MpcController{
		ID: myConfig.ControllerId(),
		Services: []mpc_controller.MpcControllerService{
			&groupService,
			&keygenService,
			&stakeService,
			//&reward.Reward{},
		},
	}

	return &controller
}

func initConfig(c *cli.Context) (logger.Logger, config.Config, storage.Storer) {
	// Parse config file
	configImpl := config.ParseConfigFromFile(c.String(configFile))

	// Create globally shared logger
	logger.DevMode = configImpl.IsDevMode()
	log := logger.Default()

	// Initialize config
	configInterface := config.InitConfig(log, configImpl)

	// Initiate storer
	storer := storage.New(log, configImpl.DatabasePath())

	return log, configInterface, storer
}
