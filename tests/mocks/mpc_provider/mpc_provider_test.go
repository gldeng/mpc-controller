package mpc_provider

import (
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	"context"
	"math/big"
	"os"
	"time"
)

type mpcProviderTestSuite struct {
	suite.Suite

	log         logger.Logger
	cChainId    *big.Int
	cPrivateKey *ecdsa.PrivateKey
	cRpcClient  *ethclient.Client
	cWsClient   *ethclient.Client

	participantPrivKeyHexs []string

	stakeAddr       common.Address
	mpcProvider     *MpcProvider
	coordinatorAddr common.Address
}

func (suite *mpcProviderTestSuite) SetupTest() {
	logger.DevMode = true

	suite.log = logger.Default()
	suite.cChainId = big.NewInt(43112)

	privateKey, _ := ethCrypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	suite.cPrivateKey = privateKey

	suite.cRpcClient = network.DefaultEthClient()
	suite.cWsClient = network.DefaultWsEthClient()

	coordinatorAddr := common.HexToAddress("0x273487EfaC011cfb62361f7b3E3763A54A03D1d3")
	suite.coordinatorAddr = coordinatorAddr

	suite.participantPrivKeyHexs = []string{
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}
}

// todo: refactor
func (suite *mpcProviderTestSuite) TestCreateGroupAndRequestKeygen() {
	//require := suite.Require()

	ctx, shutdown := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	mpcProvider := New(suite.log, suite.cChainId, &suite.coordinatorAddr, suite.cPrivateKey, suite.cRpcClient, suite.cWsClient)
	suite.mpcProvider = mpcProvider

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

		shutdown()
		return nil
	})

	<-gctx.Done()
}

func (suite *mpcProviderTestSuite) TestCreateGroup() {
	//require := suite.Require()

	ctx, shutdown := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	mpcProvider := New(suite.log, suite.cChainId, &suite.coordinatorAddr, suite.cPrivateKey, suite.cRpcClient, suite.cWsClient)
	suite.mpcProvider = mpcProvider

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
		suite.log.Debug("[][][][][] Created group", logger.Field{"groupId", groupId})

		shutdown()
		return nil
	})

	<-gctx.Done()
}

func (suite *mpcProviderTestSuite) TestRequestKeygen() {
	//require := suite.Require()

	ctx, shutdown := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	mpcProvider := New(suite.log, suite.cChainId, &suite.coordinatorAddr, suite.cPrivateKey, suite.cRpcClient, suite.cWsClient)
	suite.mpcProvider = mpcProvider

	g.Go(func() error {
		groupId := "3726383e52fd4cb603498459e8a4a15d148566a51b3f5bfbbf3cac7b61647d04"

		ethPubKeyHex, err := suite.mpcProvider.RequestKeygen(groupId)
		if err != nil {
			logger.Error("Got an error when request key generation")
			os.Exit(1)
		}
		suite.log.Debug("%%%%%%%% emitted public key", logger.Field{"publicKey", ethPubKeyHex})

		// todo: this function does not work right all the time
		stakeAddr, err := addrs.EthPubkeyHexToAddress(ethPubKeyHex)
		if err != nil {
			logger.Error("Got an error when parse address from Ethereum public key")
			os.Exit(1)
		}
		suite.stakeAddr = *stakeAddr

		suite.log.Debug("######## Generated key emitted",
			[]logger.Field{{"ethAddress", suite.stakeAddr}, {"ethPubkey", ethPubKeyHex}}...)

		shutdown()
		return nil
	})

	<-gctx.Done()
}

func TestMpcProviderTestSuite(t *testing.T) {
	suite.Run(t, new(mpcProviderTestSuite))
}
