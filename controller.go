package mpc_controller

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/contract/wrappers"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/services/group"
	"github.com/avalido/mpc-controller/storage"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

type MpcController struct {
	ID       string
	Services []MpcControllerService
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

// todo: add concrete services and necessary configs

func RunMpcController(c *cli.Context) error {
	// Initiate config
	myLogger, myConfig, myStorer := initConfig(c)

	// Build services
	privateKey := myConfig.ControllerKey()
	pubKeyBytes := myCrypto.MarshalPubkey(&privateKey.PublicKey)[1:]
	pubKeyHex := common.Bytes2Hex(pubKeyBytes)
	pubKeyHash := crypto.Keccak256Hash(pubKeyBytes)

	// Build group service
	groupService := group.Group{
		PubKeyStr:               pubKeyHex,
		PubKeyBytes:             pubKeyBytes,
		PubKeyHashStr:           pubKeyHash.Hex(),
		Logger:                  myLogger,
		CallerGetGroup:          &wrappers.MpcManagerCallerWrapper{myLogger, &myConfig.CoordinatorBoundInstance().MpcManagerCaller},
		WatcherParticipantAdded: &wrappers.MpcManagerFilterWrapper{myLogger, &myConfig.CoordinatorBoundInstance().MpcManagerFilterer},

		StorerStoreGroupInfo:       myStorer,
		StorerStoreParticipantInfo: myStorer,
	}

	// Handle graceful shutdown.
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		shutdown()
	}()

	// Run the mpc-controller
	controller := MpcController{
		Services: []MpcControllerService{
			&groupService,
			//&keygen.Keygen{},
			//&stake.Manager{},
			//&reward.Reward{},
		},
	}

	if err := controller.Run(shutdownCtx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func initConfig(c *cli.Context) (logger.Logger, config.Config, storage.Storer) {
	// Parse config file
	configImpl := config.ParseConfigFromFile(c.String(configFile))

	// Create globally shared logger
	logger.DevMode = configImpl.IsDevMode()
	log := logger.Default()

	// Initialize config
	configInterface := config.InitConfig(log, configImpl)

	// Start task manager
	//ctx, shutdown := context.WithCancel(context.Background())
	//staker := task.NewStaker(log, configInterface.CChainIssueClient(), configInterface.PChainIssueClient())
	storer := storage.New(log, configImpl.DatabasePath())

	return log, configInterface, storer
}
