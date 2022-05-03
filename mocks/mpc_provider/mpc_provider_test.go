package mpc_provider

import (
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"
	"testing"
)

type mpcProviderTestSuite struct {
	suite.Suite
}

func (suite *mpcProviderTestSuite) SetupTest() {
	logger.DevMode = true
}

func (suite *mpcProviderTestSuite) TestMpcProvider() {
	require := suite.Require()

	privateKey, err := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	require.Nil(err)

	rpcClient := network.DefaultEthClient()
	wsClient := network.DefaultWsEthClient()
	mpcProvider := New(43112, privateKey, rpcClient, wsClient)

	// Deploy coordinator contract
	_, _, err = mpcProvider.DeployContract()
	require.Nil(err)

	// Create group
	var participant_priv_keys = []string{
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}
	var pubKeys []*ecdsa.PublicKey
	for _, pubKey := range participant_priv_keys {
		privateKey, err := crypto.HexToECDSA(pubKey)
		require.Nil(err)
		pubKeys = append(pubKeys, &privateKey.PublicKey)
	}
	_, err = mpcProvider.CreateGroup(pubKeys, 1)
	require.Nil(err)
}

func TestMpcProviderTestSuite(t *testing.T) {
	suite.Run(t, new(mpcProviderTestSuite))
}
