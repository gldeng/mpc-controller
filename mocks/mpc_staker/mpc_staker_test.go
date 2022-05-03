package mpc_staker

import (
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mocks/mpc_provider"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"
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
	mpcProvider := mpc_provider.New(43112, privateKey, rpcClient, wsClient)

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

func (suite *mpcStakerTestSuite) TestMpcProvider() {
	require := suite.Require()

	// Request stake after key added
	cHttpUrl := "http://localhost:9650/ext/bc/C/rpc"
	cWebsocketUrl := "ws://127.0.0.1:9650/ext/bc/C/ws"
	nodeID := "NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5"
	mpcStaker := New(43112, suite.privateKeyHex, suite.coordinatorAddrHex, cHttpUrl, cWebsocketUrl)
	err := mpcStaker.RequestStakeAfterKeyAdded(suite.groupIdHex, nodeID, 30_000_000_000, 21)
	require.Nil(err)
}

func TestMpcStakerTestSuite(t *testing.T) {
	suite.Run(t, new(mpcStakerTestSuite))
}
