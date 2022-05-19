package mpc_staker

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mocks/mpc_provider"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math/big"
	"os"

	"testing"
)

type mpcStakerTestSuite struct {
	suite.Suite
	coordinatorAddrHex string
	groupIdHex         string
	privateKeyHex      string
}

func (suite *mpcStakerTestSuite) SetupTest() {
	logger.DevMode = true

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

	// Deploy coordinator contract
	addr, _, err := mpcProvider.DeployContract()
	if err != nil {
		logger.Error("Failed to deploy coordinator contract",
			logger.Field{"error", err})
		os.Exit(1)
	}
	suite.coordinatorAddrHex = addr.Hex()

	// Create group
	var participant_priv_keys = []string{
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}
	var pubKeys []*ecdsa.PublicKey
	for _, pubKey := range participant_priv_keys {
		privateKey, err := crypto.HexToECDSA(pubKey)
		if err != nil {
			logger.Error("Failed to parse C-Chain private key",
				logger.Field{"privateKeyHex", cPrivKeyHex},
				logger.Field{"error", err})
			os.Exit(1)
		}
		pubKeys = append(pubKeys, &privateKey.PublicKey)
	}
	groupId, err := mpcProvider.CreateGroup(pubKeys, 1)
	if err != nil {
		logger.Error("Failed to create group",
			logger.Field{"error", err})
		os.Exit(1)
	}
	suite.groupIdHex = groupId
}

func (suite *mpcStakerTestSuite) TestMpcStaker() {
	require := suite.Require()

	// Request stake after key added
	cHttpUrl := "http://localhost:9650/ext/bc/C/rpc"
	cWebsocketUrl := "ws://127.0.0.1:9650/ext/bc/C/ws"
	nodeID := "NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5"

	// Create eth rpc client
	ethRpcCli, err := ethclient.Dial(cHttpUrl)
	require.Nil(err, "Failed to connect eth rpc client")

	// Create eth ws client
	ethWsCli, err := ethclient.Dial(cWebsocketUrl)
	require.Nil(err, "Failed to connect eth ws client")

	// Convert coordinator address
	coordinatorAddr := common.HexToAddress(suite.coordinatorAddrHex)
	require.Nil(err, "Failed to parse private key")

	privateKey, err := crypto.HexToECDSA(suite.privateKeyHex)

	mpcStaker := New(logger.DefaultLogger, big.NewInt(43112), &coordinatorAddr, privateKey, ethRpcCli, ethWsCli)
	amountToStake := big.NewInt(9_000_000_000_000_000_000)
	err = mpcStaker.RequestStakeAfterKeyAdded(suite.groupIdHex, nodeID, amountToStake, 21)
	require.Nil(err)
}

func TestMpcStakerTestSuite(t *testing.T) {
	suite.Run(t, new(mpcStakerTestSuite))
}

// ----------

// todo: move this to mpc-provider
func TestMpcStaker_RequestKeyGen(t *testing.T) {
	logger.DevMode = true

	privateKeyHex := "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	require.Nil(t, err)

	// Request stake after key added
	cHttpUrl := "http://localhost:9650/ext/bc/C/rpc"
	cWebsocketUrl := "ws://127.0.0.1:9650/ext/bc/C/ws"

	// Create eth rpc client
	ethRpcCli, err := ethclient.Dial(cHttpUrl)
	require.Nil(t, err)

	// Create eth ws client
	ethWsCli, err := ethclient.Dial(cWebsocketUrl)
	require.Nil(t, err)

	// Convert coordinator address
	coordinatorAddr := common.HexToAddress("0x52C84043CD9c865236f11d9Fc9F56aa003c1f922")

	mpcStaker := New(logger.Default(), big.NewInt(43112), &coordinatorAddr, privateKey, ethRpcCli, ethWsCli)

	pubkeyHex, err := mpcStaker.requestKeygen("3726383e52fd4cb603498459e8a4a15d148566a51b3f5bfbbf3cac7b61647d04")
	require.Nil(t, err)
	fmt.Printf("Got generated public key hex: %q", pubkeyHex)
}
