package avalido_staker

import (
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/avalido/mpc-controller/utils/token"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type AvaLidoTestSuite struct {
	suite.Suite

	log         logger.Logger
	cChainId    *big.Int
	cPrivateKey *ecdsa.PrivateKey
	cRpcClient  *ethclient.Client
	cWsClient   *ethclient.Client

	avaLidoAddr *common.Address
}

func (suite *AvaLidoTestSuite) SetupTest() {
	logger.DevMode = true

	suite.log = logger.Default()
	suite.cChainId = big.NewInt(43112)

	privateKey, _ := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	suite.cPrivateKey = privateKey

	suite.cRpcClient = network.DefaultEthClient()
	suite.cWsClient = network.DefaultWsEthClient()

	avalidoAddr := common.HexToAddress("0x65f07C5D38E95C6EAc92aD1A13a0D54115915659")
	suite.avaLidoAddr = &avalidoAddr
}

func (suite *AvaLidoTestSuite) TestInitiateStake() {
	require := suite.Require()

	// Note: the stake account must at least has 25 AVAX to become a delegator.
	err := token.TransferInCChain(suite.cRpcClient, suite.cChainId, suite.cPrivateKey, suite.avaLidoAddr, big.NewInt(26_000_000_000))
	require.Nilf(err, "Failed to transfer token. error: %v", err)

	a := New(suite.log, suite.cChainId, suite.avaLidoAddr, suite.cPrivateKey, suite.cRpcClient, suite.cWsClient)
	err = a.InitiateStake(big.NewInt(25_000_000_000))
	require.Nil(err)
}

func TestAvaLidoTestSuite(t *testing.T) {
	suite.Run(t, new(AvaLidoTestSuite))
}
