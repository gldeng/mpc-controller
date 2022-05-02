package network

// Basic imports
import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type EthclientTestSuite struct {
	suite.Suite
}

func (suite *EthclientTestSuite) SetupTest() {}

func (suite *EthclientTestSuite) TestNew() {
	require := suite.Require()

	client := NewEthClient("http://localhost:9650/ext/bc/C/rpc")
	require.NotNil(client)
}

func (suite *EthclientTestSuite) TestDefault() {
	require := suite.Require()

	client1 := DefaultEthClient()
	require.NotNil(client1)

	client2 := DefaultEthClient()
	require.NotNil(client2)

	require.Equal(client2, client1)
}

func TestEthclientTestSuite(t *testing.T) {
	suite.Run(t, new(EthclientTestSuite))
}
