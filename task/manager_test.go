package task

import (
	"context"
	"crypto/ecdsa"
	//"github.com/ava-labs/avalanchego/ids"
	//"github.com/ava-labs/avalanchego/vms/components/avax"
	//"github.com/ava-labs/avalanchego/vms/platformvm"
	//"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/config"
	//"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mocks/mpc_provider"
	"github.com/avalido/mpc-controller/mocks/mpc_staker"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"math/big"
	"os"
	"testing"
	"time"
)

const (
	AVAX_ID      = "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
	CCHAIN_ID    = "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU"
	NETWORK_ID   = 12345
	CHAIN_ID     = 43112
	GAS_PER_BYTE = 1
	GAS_PER_SIG  = 1000
	GAS_FIXED    = 10000
	IMPORT_FEE   = 1000000
)

type TaskManagerTestSuite struct {
	suite.Suite
	participantPrivKeyHexs []string
	//coordinatorAddrHex     string
	coordinatorAddr *common.Address
	groupIdHex      string
	privateKeyHex   string
	mpcProvider     *mpc_provider.MpcProvider
}

// todo: transfer before doing stuff if neccessary

func (suite *TaskManagerTestSuite) SetupTest() {
	logger.DevMode = true

	suite.participantPrivKeyHexs = []string{
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}

	cPrivKeyHex := "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
	suite.privateKeyHex = cPrivKeyHex

	privateKey, err := crypto.HexToECDSA(cPrivKeyHex)
	if err != nil {
		logger.Error("Failed to parse C-Chain private key",
			logger.Field{"privateKeyHex", cPrivKeyHex},
			logger.Field{"error", err})
		os.Exit(1)
	}

	rpcClient := network.DefaultEthClient()
	wsClient := network.DefaultWsEthClient()
	mpcProvider := mpc_provider.New(logger.Default(), big.NewInt(43112), privateKey, rpcClient, wsClient)
	suite.mpcProvider = mpcProvider

	// Deploy coordinator contract
	addr, _, err := mpcProvider.DeployContract()
	if err != nil {
		logger.Error("Failed to deploy coordinator contract",
			logger.Field{"error", err})
		os.Exit(1)
	}
	suite.coordinatorAddr = addr
	//suite.coordinatorAddrHex = addr.Hex()
}

func (suite *TaskManagerTestSuite) TestTaskManagerGroup() {
	require := suite.Require()

	ctx, shutdown := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	// todo: deal with elegant shutdown
	_ = shutdown
	_ = gctx

	// Simulate creating group
	g.Go(func() error {
		time.Sleep(time.Second * 5)
		var pubKeys []*ecdsa.PublicKey
		for _, pubKey := range suite.participantPrivKeyHexs {
			privateKey, err := crypto.HexToECDSA(pubKey)
			if err != nil {
				logger.Error("Failed to parse C-Chain private key",
					logger.Field{"privateKeyHex", suite.privateKeyHex},
					logger.Field{"error", err})
				os.Exit(1)
			}
			pubKeys = append(pubKeys, &privateKey.PublicKey)
		}
		groupId, err := suite.mpcProvider.CreateGroup(pubKeys, 1)
		if err != nil {
			logger.Error("Failed to create group",
				logger.Field{"error", err})
			os.Exit(1)
		}
		suite.groupIdHex = groupId
		return nil
	})

	// Simulate request stake after key added
	g.Go(func() error {
		time.Sleep(time.Second * 10)
		nodeID := "NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg"

		privateKey, err := crypto.HexToECDSA(suite.privateKeyHex)
		require.Nil(err)

		rpcClient := network.DefaultEthClient()
		wsClient := network.DefaultWsEthClient()

		mpcStaker := mpc_staker.New(logger.Default(), big.NewInt(43112), suite.coordinatorAddr, privateKey, rpcClient, wsClient)
		for i := 1; i < 2; i++ {
			logger.Debug("RequestStakeAfterKeyAdded started////////////////////////////", logger.Field{"number", i})
			amountToStake := big.NewInt(25_000_000_000)
			err := mpcStaker.RequestStakeAfterKeyAdded(suite.groupIdHex, nodeID, amountToStake, 21)
			if err != nil {
				logger.Error("Failed to RequestStakeAfterKeyAdded",
					logger.Field{"error", err})
				os.Exit(1)
			}
			time.Sleep(time.Second * 5)
		}

		return nil
	})

	//// Simulate mpc-server, for mock
	//g.Go(func() error {
	//	mpc_server_openapi.ListenAndServe("9000")
	//	return nil
	//})

	// Do the mpc-controller stuff
	//type Arg struct {
	//	privateKey string
	//	mpc_url    string
	//}

	var configFiles = []string{
		"./testConfigs/config1.yaml",
		"./testConfigs/config2.yaml",
		"./testConfigs/config3.yaml",
	}

	for _, configFile := range configFiles {
		configImpl := config.ParseConfigFromFile(configFile)
		configImpl.SetCoordinatorAddress(suite.coordinatorAddr.Hex())
		configInterface := config.InitConfig(configImpl)

		g.Go(func() error {
			logger.DevMode = configInterface.IsDevMode()
			log := logger.Default()

			staker := NewStaker(log, configInterface.CChainIssueClient(), configInterface.PChainIssueClient())

			m, err := NewTaskManager(log, configInterface, staker)
			if err != nil {
				return errors.Wrap(err, "Failed to create task-manager for mpc-controller")
			}

			//err = m.Initialize()
			//if err != nil {
			//	return errors.Wrapf(err, "Failed to initialize task-manager %d", i)
			//}

			err = m.Start()
			if err != nil {
				return errors.Wrap(err, "Failed to start task-manager for mpc-controller")
			}
			return nil
		})
	}

	//var args = []Arg{
	//	//{privateKey: suite.participantPrivKeyHexs[0], mpc_url: "http://localhost:8001"},
	//	//{privateKey: suite.participantPrivKeyHexs[1], mpc_url: "http://localhost:8002"},
	//	//{privateKey: suite.participantPrivKeyHexs[2], mpc_url: "http://localhost:8003"},
	//	{privateKey: suite.participantPrivKeyHexs[0], mpc_url: "http://localhost:9000"},
	//	{privateKey: suite.participantPrivKeyHexs[1], mpc_url: "http://localhost:9000"},
	//	{privateKey: suite.participantPrivKeyHexs[2], mpc_url: "http://localhost:9000"},
	//}
	//
	//for i, arg := range args {
	//	i := i
	//	arg := arg
	//	g.Go(func() error {
	//		i++
	//		networkCtx, err := testNetworkContext()
	//		if err != nil {
	//			return errors.Wrap(err, "failed to build Avalanche network context")
	//		}
	//
	//		sk, err := crypto.HexToECDSA(arg.privateKey)
	//		if err != nil {
	//			return errors.Wrapf(err, "failed to parse private key from %q", arg.privateKey)
	//		}
	//
	//		mpcClient, err := core.NewMpcClient(arg.mpc_url)
	//		if err != nil {
	//			return errors.Wrapf(err, "failed to build mpc-client %d with mpc-url %q", i, arg.mpc_url)
	//		}
	//
	//		//mpcClient := mpc_client.New(3, 1)
	//
	//		cChainIssueUrl := "http://localhost:9650"
	//		pChainIssueUrl := "http://localhost:9650"
	//		// Create C-Chain issue client
	//		cChainIssueCli := evm.NewClient(cChainIssueUrl, "C")
	//
	//		// Create P-Chain issue client
	//		pChainIssueCli := platformvm.NewClient(pChainIssueUrl)
	//
	//
	//
	//		staker := NewStaker(logger.DefaultLogger, cChainIssueCli, pChainIssueCli)
	//		manager, err := NewTaskManager(logger.Default(), config, staker)
	//		if err != nil {
	//			return errors.Wrapf(err, "Failed to build task-manager %d", i)
	//		}
	//		logger.Debug("A task manager created", logger.Field{"taskMangerNum", i})
	//
	//		//err = manager.Initialize()
	//		//if err != nil {
	//		//	return errors.Wrapf(err, "Failed to initialize task-manager %d", i)
	//		//}
	//
	//		logger.Debug("Started a task manager",
	//			logger.Field{"managerID", i},
	//			logger.Field{"privateKey", arg.privateKey},
	//			logger.Field{"mpcURL", arg.mpc_url},
	//			logger.Field{"contractAddress", coordinatorAddr})
	//		err = manager.Start()
	//		if err != nil {
	//			return errors.Wrapf(err, "Failed to start task-manager %d", i)
	//		}
	//		return nil
	//	})
	//}

	err := g.Wait()
	require.Nilf(err, "ERROR STACK: %+v", errors.WithStack(err))
}

func TestTaskManagerTestSuite(t *testing.T) {
	suite.Run(t, new(TaskManagerTestSuite))
}
