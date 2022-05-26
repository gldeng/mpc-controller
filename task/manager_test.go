package task

import (
	"context"
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/mocks/avalido_staker"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/token"

	//"github.com/ava-labs/avalanchego/ids"
	//"github.com/ava-labs/avalanchego/vms/components/avax"
	//"github.com/ava-labs/avalanchego/vms/platformvm"
	//"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/config"
	//"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mocks/mpc_provider"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/ethclient"
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

	log         logger.Logger
	cChainId    *big.Int
	cPrivateKey *ecdsa.PrivateKey
	cRpcClient  *ethclient.Client
	cWsClient   *ethclient.Client

	participantPrivKeyHexs []string
	//coordinatorAddrHex     string
	//groupIdHex string
	//privateKeyHex   string

	stakeAddr       common.Address
	mpcProvider     *mpc_provider.MpcProvider
	coordinatorAddr common.Address

	avaLidoStaker *avalido_staker.AvaLidoStaker
	avaLidoAddr   common.Address
}

// todo: transfer before doing stuff if neccessary

func (suite *TaskManagerTestSuite) SetupTest() {
	logger.DevMode = true

	suite.log = logger.Default()
	suite.cChainId = big.NewInt(43112)

	privateKey, _ := ethCrypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	suite.cPrivateKey = privateKey

	suite.cRpcClient = network.DefaultEthClient()
	suite.cWsClient = network.DefaultWsEthClient()

	avalidoAddr := common.HexToAddress("0xA4cD3b0Eb6E5Ab5d8CE4065BcCD70040ADAB1F00")
	suite.avaLidoAddr = avalidoAddr
	// ----

	suite.participantPrivKeyHexs = []string{
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}
}

func (suite *TaskManagerTestSuite) TestTaskManagerGroup() {
	require := suite.Require()

	ctx, shutdown := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	// todo: deal with elegant shutdown
	_ = shutdown
	_ = gctx

	// Deploy coordinator contract
	coordinatorAdd, _, err := mpc_provider.DeployMpcManager(suite.log, suite.cChainId, suite.cRpcClient, suite.cPrivateKey)
	require.Nilf(err, "error: %v", err)
	suite.log.Info("Deployed mpc coordinator", logger.Field{"coordinatorAddress", coordinatorAdd.Hex()})
	suite.coordinatorAddr = *coordinatorAdd
	suite.log.Debug("######## Coordinator contract address", logger.Field{"address", suite.coordinatorAddr.Hex()})
	mpcProvider := mpc_provider.New(suite.log, suite.cChainId, &suite.coordinatorAddr, suite.cPrivateKey, suite.cRpcClient, suite.cWsClient)
	suite.mpcProvider = mpcProvider

	// Deploy AvaLido contract
	avaLidoAddr, _, err := avalido_staker.DeployAvaLido(suite.log, suite.cChainId, suite.cRpcClient, suite.cPrivateKey, &suite.coordinatorAddr)
	require.Nilf(err, "error: %v", err)
	suite.avaLidoAddr = *avaLidoAddr
	suite.log.Debug("######## AvaLido contract address", logger.Field{"address", suite.avaLidoAddr.Hex()})
	avalidoStaker := avalido_staker.New(suite.log, suite.cChainId, avaLidoAddr, suite.cPrivateKey, suite.cRpcClient, suite.cWsClient)
	suite.avaLidoStaker = avalidoStaker

	// Simulate creating group and request key generating
	g.Go(func() error {
		time.Sleep(time.Second * 5)
		var pubKeys []*ecdsa.PublicKey
		for _, pubKey := range suite.participantPrivKeyHexs {
			privateKey, err := ethCrypto.HexToECDSA(pubKey)
			if err != nil {
				suite.log.Error("Failed to parse C-Chain private key",
					logger.Field{"error", err})
				os.Exit(1)
			}
			pubKeys = append(pubKeys, &privateKey.PublicKey)
		}
		groupId, err := suite.mpcProvider.CreateGroup(pubKeys, 1)
		if err != nil {
			suite.log.Error("Failed to create group",
				logger.Field{"error", err})
			os.Exit(1)
		}

		pubKeyHex, err := suite.mpcProvider.RequestKeygen(groupId)
		if err != nil {
			logger.Error("Got an error when request key generation")
			os.Exit(1)
		}

		pubKey, err := crypto.UnmarshalPubKeyHex(pubKeyHex)
		if err != nil {
			logger.Error("Got an error when unmarshal public key")
			os.Exit(1)
		}

		stakeAddr := ethCrypto.PubkeyToAddress(*pubKey)
		suite.stakeAddr = stakeAddr
		suite.log.Debug("######## Account to stake", logger.Field{"address", suite.stakeAddr})
		return nil
	})

	// Simulate initiate stake by AvaLido staker after key generated
	// TODO: add code to check participant balance
	g.Go(func() error {
		time.Sleep(time.Second * 15)
		for i := 1; i < 2; i++ {
			// Check balance before transfer
			bl, _ := suite.cRpcClient.BalanceAt(context.Background(), suite.avaLidoAddr, nil)
			suite.log.Debug("$$$$$$$$$0 Balance of AvaLido address before transfer", []logger.Field{{"address", suite.avaLidoAddr}, {"balance", bl.Uint64()}}...)

			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.coordinatorAddr, nil)
			suite.log.Debug("$$$$$$$$$0 Balance of Coordinator address before transfer", []logger.Field{{"address", suite.coordinatorAddr}, {"balance", bl.Uint64()}}...)

			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.stakeAddr, nil)
			suite.log.Debug("$$$$$$$$$0 Balance of stake address before transfer", []logger.Field{{"address", suite.stakeAddr.Hex()}, {"balance", bl.Uint64()}}...)

			// Transfer to AvaLido smart contract to make sure there is enough balance to be deducted for initiating stake
			// Including gas fee
			err := token.TransferInCChain(suite.cRpcClient, suite.cChainId, suite.cPrivateKey, &suite.avaLidoAddr, big.NewInt(1_000_000_000_000_000_000))
			require.Nilf(err, "Failed to transfer token. error: %v", err)

			time.Sleep(time.Second * 10)

			// Transfer gas fee to Coordinator smart contract
			err = token.TransferInCChain(suite.cRpcClient, suite.cChainId, suite.cPrivateKey, &suite.coordinatorAddr, big.NewInt(1_000_000_000_000_000_000))
			require.Nilf(err, "Failed to transfer token. error: %v", err)

			time.Sleep(time.Second * 10)

			// Transfer gas fee to stake address
			err = token.TransferInCChain(suite.cRpcClient, suite.cChainId, suite.cPrivateKey, &suite.stakeAddr, big.NewInt(1_000_000_000_000_000_000))
			require.Nilf(err, "Failed to transfer token. error: %v", err)

			time.Sleep(time.Second * 10)

			// Check balance before initiating stake
			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.avaLidoAddr, nil)
			suite.log.Debug("$$$$$$$$$1 Balance of AvaLido address before initiating stake", []logger.Field{{"address", suite.avaLidoAddr}, {"balance", bl.Uint64()}}...)

			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.coordinatorAddr, nil)
			suite.log.Debug("$$$$$$$$$1 Balance of Coordinator address before initiating stake", []logger.Field{{"address", suite.coordinatorAddr}, {"balance", bl.Uint64()}}...)

			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.stakeAddr, nil)
			suite.log.Debug("$$$$$$$$$1 Balance of stake address before initiating stake", []logger.Field{{"address", suite.stakeAddr.Hex()}, {"balance", bl.Uint64()}}...)

			amountToStake := big.NewInt(25_000_000_000)
			err = suite.avaLidoStaker.InitiateStake(amountToStake)
			if err != nil {
				logger.Error("Failed to initiate stake",
					logger.Field{"error", err})
				os.Exit(1)
			}

			// Check balance after initiating stake
			time.Sleep(time.Second * 15)
			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.avaLidoAddr, nil)
			suite.log.Debug("$$$$$$$$$2 Balance of AvaLido address after initiating stake", []logger.Field{{"address", suite.avaLidoAddr}, {"balance", bl.Uint64()}}...)

			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.coordinatorAddr, nil)
			suite.log.Debug("$$$$$$$$$2 Balance of Coordinator address after initiating stake", []logger.Field{{"address", suite.coordinatorAddr}, {"balance", bl.Uint64()}}...)

			bl, _ = suite.cRpcClient.BalanceAt(context.Background(), suite.stakeAddr, nil)
			suite.log.Debug("$$$$$$$$$2 Balance of stake address after initiating stake", []logger.Field{{"address", suite.stakeAddr.Hex()}, {"balance", bl.Uint64()}}...)
		}

		return nil
	})

	// Start mpc controllers
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

			storer := storage.New(log, configImpl.DatabasePath())
			m, err := NewTaskManager(log, configInterface, storer, staker)
			if err != nil {
				return errors.Wrap(err, "Failed to create task-manager for mpc-controller")
			}

			err = m.Start()
			if err != nil {
				return errors.Wrap(err, "Failed to start task-manager for mpc-controller")
			}
			return nil
		})
	}
	err = g.Wait()
	require.Nilf(err, "ERROR STACK: %+v", errors.WithStack(err))
}

func TestTaskManagerTestSuite(t *testing.T) {
	suite.Run(t, new(TaskManagerTestSuite))
}
